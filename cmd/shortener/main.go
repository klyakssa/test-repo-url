package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/handler"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/router"
	"github.com/klyakssa/test-repo-url/internal/service/uuidservice"
	"github.com/klyakssa/test-repo-url/internal/uuidstorage"
)

func main() {

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

	h := handler.NewMyHandler(log, userService, config)
	r.Middleware(log.WithLogging())
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

	go func() {
		if err := r.Run(config.WebConfig.HostPort); err != nil {
			errChan <- err
		}
	}()

	defer func() {
		cancel()
		if err := userService.Close(); err != nil {
			log.Error(err)
		} else {
			log.Info("Shutting down gracefully")
		}
	}()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		select {
		case <-sigChan:
			log.Info("Shutdown signal received")
			cancel()
		case <-ctx.Done():
		}
	}()

	select {
	case err := <-errChan:
		log.Error("Application terminated with error: %v", err)
		cancel()
	case <-ctx.Done():
		log.Info("Application terminated gracefully")
	}
}
