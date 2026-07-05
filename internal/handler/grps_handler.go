package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
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
	grpcServer *grpc.Server
	config     *config.Config
}

func NewShortenerServer(
	svc repository.UserService,
	subscriber *audit.Audit,
	logger *logger.MyLogger,
	config *config.Config,
) *ShortenerServer {
	return &ShortenerServer{
		service:    svc,
		subscriber: subscriber,
		logger:     logger,
		config:     config,
	}
}

func (s *ShortenerServer) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		s.logger.Debug("AuthInterceptor", zap.String("method", info.FullMethod))

		key := sha256.Sum256([]byte(s.config.WebConfig.Secret))

		aesblock, err := aes.NewCipher(key[:])
		if err != nil {
			s.logger.Error(err)
			return nil, status.Error(codes.Internal, "internal error")
		}

		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			s.logger.Error(err)
			return nil, status.Error(codes.Internal, "internal error")
		}

		nonce := key[len(key)-aesgcm.NonceSize():]

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			uuid := uuid.NewString()
			userID := hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil))
			ctx = context.WithValue(ctx, UserIDKey, userID)

			s.logger.Debug("New user created", zap.String("user_id", userID))

			resp, err := handler(ctx, req)
			if err != nil {
				s.logger.Error("Failed to handle request", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to handle request")
			}
			return resp, nil
		}

		token := authHeader[0]
		if !strings.HasPrefix(token, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		encryptedUserID := strings.TrimPrefix(token, "Bearer ")

		data, err := hex.DecodeString(encryptedUserID)
		if err != nil {
			s.logger.Debug("Failed to decode user ID", zap.Error(err))
			uuid := uuid.NewString()
			userID := hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil))
			ctx = context.WithValue(ctx, UserIDKey, userID)

			s.logger.Debug("New user created", zap.String("user_id", userID))

			resp, err := handler(ctx, req)
			if err != nil {
				s.logger.Error("Failed to handle request", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to handle request")
			}
			return resp, nil
		}

		userID, err := aesgcm.Open(nil, nonce, data, nil)
		if err != nil {
			uuid := uuid.NewString()
			userID := hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil))
			ctx = context.WithValue(ctx, UserIDKey, userID)

			s.logger.Debug("New user created", zap.String("user_id", userID))

			resp, err := handler(ctx, req)
			if err != nil {
				s.logger.Error("Failed to handle request", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to handle request")
			}
			return resp, nil
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)
		s.logger.Debug("User authenticated", zap.String("user_id", string(userID)))

		return handler(ctx, req)
	}
}

// ShortenURL - соответствует POST /api/shorten
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	s.logger.Debug("ShortenURL gRPC called",
		zap.String("url", req.GetUrl()),
	)

	// Получаем userID из metadata
	userID, err := s.getUserIDFromMetadata(ctx)
	if err != nil {
		userID = "unknown"
		s.logger.Warn("Failed to get userID from metadata", zap.Error(err))
	}

	// Используем существующий сервис
	shortURL, err := s.service.Shorten(ctx, model.CreateShortURLInput{
		OriginalURL: req.GetUrl(),
		UserID:      userID,
	})
	if err != nil {
		s.logger.Error("Failed to shorten URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	s.logger.Debug("ShortenURL gRPC success",
		zap.String("original", req.GetUrl()),
		zap.String("short", shortURL),
		zap.String("user_id", userID),
	)

	// Асинхронная запись в аудит (как в HTTP-хендлере)
	go s.subscriber.Subscribe(&model.AuditEntry{
		Action: "shorten",
		UserID: userID,
		URL:    req.GetUrl(),
	})

	return pb.URLShortenResponse_builder{Result: shortURL}.Build(), nil
}

// ExpandURL - соответствует GET /<id>
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	s.logger.Debug("ExpandURL gRPC called",
		zap.String("id", req.GetId()),
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
		UUID:   req.GetId(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "URL not found")
		}
		// Проверяем на ошибку удаления
		if errors.Is(err, postgres.ErrURLDeleted) {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}

		s.logger.Error("Failed to expand URL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to expand URL: %v", err)
	}

	s.logger.Debug("ExpandURL gRPC success",
		zap.String("id", req.GetId()),
		zap.String("original", originalURL),
		zap.String("user_id", userID),
	)

	// Асинхронная запись в аудит (как в HTTP-хендлере)
	go s.subscriber.Subscribe(&model.AuditEntry{
		Action: "follow",
		UserID: userID,
		URL:    originalURL,
	})

	return pb.URLExpandResponse_builder{Result: originalURL}.Build(), nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *pb.ListUserURLsRequest) (*pb.UserURLsResponse, error) {
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
		pbUrls = append(pbUrls, pb.URLData_builder{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}.Build())
	}

	s.logger.Debug("ListUserURLs gRPC success",
		zap.String("user_id", userID),
		zap.Int("count", len(pbUrls)),
	)

	return pb.UserURLsResponse_builder{Urls: pbUrls}.Build(), nil
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

func (s *ShortenerServer) setupTLS() (grpc.ServerOption, error) {
	if !s.config.WebConfig.EnableHTTPS {
		return nil, nil
	}

	if _, err := os.Stat(s.config.WebConfig.CertFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate file not found: %s", s.config.WebConfig.CertFile)
	}
	if _, err := os.Stat(s.config.WebConfig.KeyFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file not found: %s", s.config.WebConfig.KeyFile)
	}

	cert, err := tls.LoadX509KeyPair(s.config.WebConfig.CertFile, s.config.WebConfig.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	creds := credentials.NewTLS(tlsConfig)

	return grpc.Creds(creds), nil
}

func (s *ShortenerServer) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	tlsOpts, err := s.setupTLS()
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}

	s.grpcServer = grpc.NewServer(tlsOpts, grpc.ChainUnaryInterceptor(s.AuthInterceptor()))

	pb.RegisterShortenerServiceServer(s.grpcServer, s)

	s.logger.Info("Starting gRPC server", zap.String("address", addr))
	return s.grpcServer.Serve(lis)
}

func (s *ShortenerServer) Close() {
	s.grpcServer.GracefulStop()
}
