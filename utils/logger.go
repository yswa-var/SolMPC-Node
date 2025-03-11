package utils

import (
	mpc "tilt-valid/internal/mpc"

	"go.uber.org/zap"
)

// Logger interface for logging messages.
func Logger(id string, testName string) mpc.Logger {
	logConfig := zap.NewDevelopmentConfig()
	logger, _ := logConfig.Build()
	logger = logger.With(zap.String("t", testName)).With(zap.String("id", id))
	return logger.Sugar()
}
