package main

import (
	"os"
	"os/signal"

	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/handler"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/router"
	"github.com/klyakssa/test-repo-url/internal/service/fileservice"
	"github.com/klyakssa/test-repo-url/internal/service/uuidservice"
)

func main() {

	config := config.InitFlagConfig()
	log := logger.NewLogger()
	r := router.NewMyRouter(config)

	fs := fileservice.New(config)
	uuid := uuidservice.New()
	data, err := fs.Load()
	if err != nil {
		panic(err)
	}
	err = uuid.Save(data)
	if err != nil {
		panic(err)
	}

	h := handler.NewMyHandler(config, log, fs, uuid)
	r.Middleware(log.WithLogging())
	r.Middleware(h.GzipMiddleware())
	r.GET("/:uuid", h.UnshortenHandler)
	r.POST("/", h.ShortenHandler)
	r.POST("/api/shorten", h.NewShortenHandler)
	go func() {
		if err := r.Run(config.WebConfig.HostPort); err != nil {
			panic(err)
		}
	}()
	defer func() {
		if err := fs.Save(uuid.Load()); err != nil {
			log.Logger.Error(err)
		}
		if err := fs.Close(); err != nil {
			log.Logger.Error(err)
		}
		log.Logger.Info("Shutting down gracefully")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
