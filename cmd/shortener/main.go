package main

import (
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

	config := config.InitFlagConfig()
	log := logger.NewLogger()
	r := router.NewMyRouter(config)

	db, err := postgres.NewPostgresStorage(config, log)
	if err != nil {
		log.Error(err)
	}

	var userService *uuidservice.UUIDService
	if err != nil {
		userService = uuidservice.New(uuidstorage.New(config))
	} else {
		userService = uuidservice.New(db)
	}

	h := handler.NewMyHandler(log, userService)
	r.Middleware(log.WithLogging())
	r.Middleware(h.GzipMiddleware())
	r.Middleware(h.ErrorMiddleware())

	r.GET("/ping", h.PingPostgresHandler)
	r.GET("/:uuid", h.UnshortenHandler)
	r.POST("/", h.ShortenHandler)
	v1 := r.Group("/api/shorten")
	v1.POST("/", h.NewShortenHandler)
	v1.POST("/batch", h.BatchHandler)

	go func() {
		if err := r.Run(config.WebConfig.HostPort); err != nil {
			log.Error(err)
			panic(err)
		}
	}()

	defer func() {
		if err := userService.Close(); err != nil {
			log.Error(err)
		} else {
			log.Info("Shutting down gracefully")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
