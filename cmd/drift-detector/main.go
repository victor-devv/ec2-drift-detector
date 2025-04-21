package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// close context on graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.New("")
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(os.Stdout)
	logLevel, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level specified, defaulting to info")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	return nil
}
