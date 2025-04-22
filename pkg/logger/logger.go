package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func New(config *config.Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level specified, defaulting to info")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	return logger
}
