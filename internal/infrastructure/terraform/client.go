package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// Client provides access to Terraform configuration and state
type Client struct {
	stateParser *StateParser
	hclParser   *HCLParser
	logger      *logging.Logger
	stateFile   string
	hclDir      string
	useHCL      bool
}

// ClientConfig holds configuration for the Terraform client
type ClientConfig struct {
	StateFile string
	HCLDir    string
	UseHCL    bool
}

// NewClient creates a new Terraform client
func NewClient(cfg ClientConfig, logger *logging.Logger) (*Client, error) {
	logger = logger.WithField("component", "terraform-client")

	// Validate configuration
	if cfg.UseHCL {
		if cfg.HCLDir == "" {
			return nil, errors.NewValidationError("HCL directory must be specified when UseHCL is true")
		}

		// Check if the directory exists
		info, err := os.Stat(cfg.HCLDir)
		if err != nil {
			return nil, errors.NewOperationalError(fmt.Sprintf("HCL directory %s does not exist", cfg.HCLDir), err)
		}

		if !info.IsDir() {
			return nil, errors.NewValidationError(fmt.Sprintf("%s is not a directory", cfg.HCLDir))
		}
	} else {
		if cfg.StateFile == "" {
			return nil, errors.NewValidationError("State file must be specified when UseHCL is false")
		}

		// Check if the file exists
		_, err := os.Stat(cfg.StateFile)
		if err != nil {
			return nil, errors.NewOperationalError(fmt.Sprintf("State file %s does not exist", cfg.StateFile), err)
		}
	}

	return &Client{
		stateParser: NewStateParser(logger),
		hclParser:   NewHCLParser(logger),
		logger:      logger,
		stateFile:   cfg.StateFile,
		hclDir:      cfg.HCLDir,
		useHCL:      cfg.UseHCL,
	}, nil
}

// GetInstance retrieves instance configuration by ID
func (c *Client) GetInstance(ctx context.Context, instanceID string) (*model.Instance, error) {
	c.logger.Info(fmt.Sprintf("Retrieving instance %s from Terraform", instanceID))

	if c.useHCL {
		// When using HCL, we can't look up by instance ID directly
		// since the ID is only known after Terraform applies the configuration
		// Instead, we get all instances and try to match by name
		instances, err := c.ListInstances(ctx)
		if err != nil {
			return nil, err
		}

		for _, instance := range instances {
			// Try to match by ID (which might be resource name in HCL mode)
			if instance.ID == instanceID {
				return instance, nil
			}

			// Try to match by any attribute that might contain the instance ID
			if resourceName, ok := instance.Attributes["resource_name"].(string); ok && resourceName == instanceID {
				return instance, nil
			}
		}

		return nil, errors.NewNotFoundError("EC2 Instance", instanceID)
	} else {
		return c.stateParser.GetInstanceByIDFromStateFile(ctx, c.stateFile, instanceID)
	}
}

// ListInstances retrieves all available instances
func (c *Client) ListInstances(ctx context.Context) ([]*model.Instance, error) {
	c.logger.Info("Listing instances from Terraform")

	if c.useHCL {
		return c.hclParser.ParseHCLDir(ctx, c.hclDir)
	} else {
		return c.stateParser.GetInstancesFromStateFile(ctx, c.stateFile)
	}
}

// GetSourceType returns the source type for this client
func (c *Client) GetSourceType() model.ResourceOrigin {
	return model.OriginTerraform
}

// IsUsingHCL returns whether the client is using HCL or state files
func (c *Client) IsUsingHCL() bool {
	return c.useHCL
}

// GetStateFile returns the state file path
func (c *Client) GetStateFile() string {
	return c.stateFile
}

// GetHCLDir returns the HCL directory path
func (c *Client) GetHCLDir() string {
	return c.hclDir
}
