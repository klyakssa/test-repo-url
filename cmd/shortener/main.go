package main

import (
	"context"
	"errors"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/handler"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/router"
	"github.com/klyakssa/test-repo-url/internal/service/uuidservice"
	"github.com/klyakssa/test-repo-url/internal/uuidstorage"
	"github.com/klyakssa/test-repo-url/pkg/audit"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	config := config.InitFlagConfig()
	log := logger.NewLogger()
	r := router.NewMyRouter(config)

	db, err := postgres.NewPostgresStorage(config, log)
	if err != nil {
		log.Error(err)
		if !errors.Is(err, postgres.ErrNoConnectionString) {
			panic(err)
		}
	}

	var userService *uuidservice.UUIDService
	if err != nil {
		userService = uuidservice.New(ctx, uuidstorage.New(config), log)
	} else {
		userService = uuidservice.New(ctx, db, log)
	}

	audit := audit.NewAudit(config.Audit.AuditFile, config.Audit.AuditURL)

	grpcServer := handler.NewShortenerServer(
		userService,
		audit,
		log,
		config,
	)

	r = initRoutes(log, userService, config, r, audit)

	go func() {
		if err := r.Run(ctx, &config.WebConfig); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := grpcServer.Run(config.GRPCConfig.GRPCPort); err != nil {
			errChan <- err
		}
	}()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

		select {
		case <-sigChan:
			log.Info("Shutdown signal received")
			signal.Stop(sigChan)
			cancel()
		case <-ctx.Done():
			signal.Stop(sigChan)
		}
	}()

	select {
	case err := <-errChan:
		log.Error("Application terminated with error: %v", err)
		cancel()
	case <-ctx.Done():
		grpcServer.Close()
		if err := userService.Close(); err != nil {
			log.Error(err)
		} else {
			log.Info("Shutting down gracefully")
		}
		log.Info("Application terminated gracefully")
	}
}

func initRoutes(log *logger.MyLogger, userService *uuidservice.UUIDService, config *config.Config, r *router.MyRouter, audit *audit.Audit) *router.MyRouter {
	h := handler.New(log, userService, config, audit)
	r.Middleware(h.WithLogging())
	r.Middleware(h.SecretMiddleware())
	r.Middleware(h.GzipMiddleware())
	r.Middleware(h.ErrorMiddleware())

	r.SGET("/ping", h.PingPostgresHandler)
	r.SGET("/:uuid", h.UnshortenHandler)
	r.SPOST("/", h.ShortenHandler)
	v1 := r.Group("/api/shorten")
	v1.SPOST("", h.NewShortenHandler)
	v1.SPOST("/batch", h.BatchHandler)
	v2 := r.Group("/api/user")
	v2.GET("/urls", h.GetUrlsHandler())
	v2.DELETE("/urls", h.DeleteUrlsHandler())
	v3 := r.Group("/api/internal")
	if config.WebConfig.TrustedSubnet.String() != "" {
		v3.Middleware(h.TrustedSubnetMiddleware())
	}
	v3.GET("/stats", h.StatsHandler())

	return r
}
