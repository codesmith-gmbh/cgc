package cgclog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MustSugaredLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}

func SyncSugaredLogger(logger *zap.SugaredLogger) {
	if err := logger.Sync(); err != nil {
		fmt.Printf("could not sync sugared logger: %v", err)
	}
}
