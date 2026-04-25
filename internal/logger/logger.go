package logger

import (
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
