package pkginit

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	var err error
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
	defer Logger.Sync()

	Logger.Debug("Logger initialized")
}
