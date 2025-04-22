/*
Parses a standard Terraform .tfstate file or HCL configuration.

Extracts EC2 instance blocks into internal EC2Instance models.

Designed for extensibility to .tf HCL parsing.
*/
package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// Parser interface defines the contract for all Terraform parsers
type Parser interface {
	// GetEC2Instances returns all EC2 instances defined in Terraform
	GetEC2Instances(ctx context.Context) ([]models.EC2Instance, error)
}

// BaseParser provides common functionality for all parsers
type BaseParser struct {
	log *logrus.Logger
}

// GetParser returns the appropriate parser based on the file extension
func GetParser(log *logrus.Logger, path string) (Parser, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if fileInfo.IsDir() {
		return NewHCLParser(path, log), nil
	} else {
		ext := filepath.Ext(path)

		switch ext {
		case ".json", ".tfstate":
			return NewStateParser(path, log), nil
		case ".tf", ".hcl":
			return NewHCLParser(path, log), nil
		default:
			return nil, fmt.Errorf("unsupported file extension: %s", ext)
		}
	}
}
