package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	instance *zap.Logger
	once     sync.Once
)

func GetLogger() *zap.Logger {
	once.Do(func() {
		logger, _ := zap.NewProduction()
		defer logger.Sync()
		instance = logger
	})
	return instance
}
