// internal/terraform/parser.go
package terraform

import (
	"context"
	"fmt"
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
