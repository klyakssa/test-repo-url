package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/model"
	pb "github.com/klyakssa/test-repo-url/internal/proto"
	"github.com/klyakssa/test-repo-url/internal/repository"
	"github.com/klyakssa/test-repo-url/pkg/audit"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	service    repository.UserService
	subscriber *audit.Audit
	logger     *logger.MyLogger
	baseURL    string
	grpcServer *grpc.Server
}

func NewShortenerServer(
	svc repository.UserService,
	subscriber *audit.Audit,
	logger *logger.MyLogger,
	baseURL string,
) *ShortenerServer {
	return &ShortenerServer{
		service:    svc,
		subscriber: subscriber,
		logger:     logger,
		baseURL:    baseURL,
	}
}

// ShortenURL - соответствует POST /api/shorten
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	s.logger.Debug("ShortenURL gRPC called",
		zap.String("url", req.Url),
	)

	// Получаем userID из metadata
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		userID = "unknown"
		s.logger.Warn("Failed to get userID from metadata", zap.Error(err))
	}

	// Используем существующий сервис
	shortURL, err := s.service.Shorten(ctx, model.CreateShortURLInput{
		OriginalURL: req.Url,
		UserID:      userID,
	})
	if err != nil {
		s.logger.Error("Failed to shorten URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	s.logger.Debug("ShortenURL gRPC success",
		zap.String("original", req.Url),
		zap.String("short", shortURL),
		zap.String("user_id", userID),
	)

	// Асинхронная запись в аудит (как в HTTP-хендлере)
	go s.subscriber.Subscribe(&model.AuditEntry{
		Action: "shorten",
		UserID: userID,
		URL:    req.Url,
	})

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL - соответствует GET /<id>
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	s.logger.Debug("ExpandURL gRPC called",
		zap.String("id", req.Id),
	)

	// Получаем userID из metadata
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		userID = "unknown"
		s.logger.Warn("Failed to get userID from metadata", zap.Error(err))
	}

	// Используем существующий сервис
	originalURL, err := s.service.Unshorten(ctx, model.GetShortURLInput{
		UserID: userID,
		UUID:   req.Id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "URL not found")
		}
		// Проверяем на ошибку удаления
		if strings.Contains(err.Error(), "deleted") {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}
		s.logger.Error("Failed to expand URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to expand URL: %v", err)
	}

	s.logger.Debug("ExpandURL gRPC success",
		zap.String("id", req.Id),
		zap.String("original", originalURL),
		zap.String("user_id", userID),
	)

	// Асинхронная запись в аудит (как в HTTP-хендлере)
	go s.subscriber.Subscribe(&model.AuditEntry{
		Action: "follow",
		UserID: userID,
		URL:    originalURL,
	})

	return &pb.URLExpandResponse{Result: originalURL}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	s.logger.Debug("ListUserURLs gRPC called")

	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		s.logger.Error("Failed to get userID from metadata", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "authentication required: %v", err)
	}

	urls, err := s.service.GetUrlsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	pbUrls := make([]*pb.URLData, 0, len(urls))
	for _, url := range urls {
		pbUrls = append(pbUrls, &pb.URLData{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	s.logger.Debug("ListUserURLs gRPC success",
		zap.String("user_id", userID),
		zap.Int("count", len(pbUrls)),
	)

	return &pb.UserURLsResponse{Urls: pbUrls}, nil
}

func (s *ShortenerServer) getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("missing authorization header")
	}

	auth := authHeaders[0]
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	userID := parts[1]
	if userID == "" {
		return "", errors.New("empty user ID")
	}

	return userID, nil
}

func (s *ShortenerServer) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = grpc.NewServer()

	pb.RegisterShortenerServiceServer(s.grpcServer, s)

	s.logger.Info("Starting gRPC server", zap.String("address", addr))
	return s.grpcServer.Serve(lis)
}

func (s *ShortenerServer) Close() {
	s.grpcServer.GracefulStop()
}
