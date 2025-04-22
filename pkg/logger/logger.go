/*
Package Logger provides centralized log configuration via logrus.

Verbose mode toggle via CLI or env.
*/
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
		logger.Warnf("Invalid log level specified (%s), defaulting to info", config.Log.Level)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return logger
}
