package factory

import (
	"context"
	"strings"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/terraform"
)

// InstanceProviderFactory creates instance providers
type InstanceProviderFactory struct {
	logger *logging.Logger
}

// NewInstanceProviderFactory creates a new instance provider factory
func NewInstanceProviderFactory(logger *logging.Logger) *InstanceProviderFactory {
	return &InstanceProviderFactory{
		logger: logger,
	}
}

// CreateAWSProvider creates an AWS instance provider
func (f *InstanceProviderFactory) CreateAWSProvider(ctx context.Context, cfg *config.Config) (service.InstanceProvider, error) {
	// Create AWS client
	env := cfg.GetEnv()
	awsClient, err := aws.NewClient(context.Background(), aws.ClientConfig{
		Region:        cfg.GetAWSRegion(),
		Profile:       cfg.GetAWSProfile(),
		Endpoint:      cfg.GetAWSEndpoint(),
		AccessKey:     cfg.GetAWSAccessKeyID(),
		SecretKey:     cfg.GetAWSSecretAccessKey(),
		UseLocalstack: strings.ToLower(env) == "dev" || strings.ToLower(env) == "development",
	}, f.logger)
	if err != nil {
		return nil, err
	}

	// Create EC2 service
	ec2Service := aws.NewEC2Service(f.logger, awsClient)
	f.logger.Info("AWS provider initialized")
	return ec2Service, nil
}

// CreateTerraformProvider creates a Terraform instance provider
func (f *InstanceProviderFactory) CreateTerraformProvider(cfg *config.Config) (service.InstanceProvider, error) {
	// Create Terraform client
	terraformClient, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: cfg.GetStateFile(),
		HCLDir:    cfg.GetHCLDir(),
		UseHCL:    cfg.GetUseHCL(),
	}, f.logger)
	if err != nil {
		return nil, err
	}
	f.logger.Info("Terraform provider initialized")
	return terraformClient, nil
}
