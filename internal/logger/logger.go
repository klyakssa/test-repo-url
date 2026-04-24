package logger

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MyLogger struct{ *zap.SugaredLogger }

func NewLogger() *MyLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return &MyLogger{logger.Sugar()}
}

func (l *MyLogger) WithLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Request = c.Request.Clone(context.WithValue(c.Request.Context(), "start_time", start.Unix()))

		c.Next()
		l.Infow("request", "method", c.Request.Method, "path", c.Request.URL.Path, "duration", time.Since(start).Seconds())
		l.Infow("response", "status", c.Writer.Status(), "size", c.Writer.Size())
	}
}
