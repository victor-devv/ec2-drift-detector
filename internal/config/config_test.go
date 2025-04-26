package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
)

func TestConfigValidate(t *testing.T) {
	// Test case 1: Valid configuration
	cfg := &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err := cfg.Validate()
	assert.NoError(t, err)

	// Test case 2: Missing AWS region
	cfg = &Config{}
	cfg.AWS.Region = ""
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 3: Missing Terraform state file with UseHCL=false
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = ""
	cfg.Terraform.UseHCL = false
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 4: Missing HCL dir with UseHCL=true
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.HCLDir = ""
	cfg.Terraform.UseHCL = true
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 5: Missing attributes
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 6: Invalid source of truth
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "invalid"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 7: Invalid parallel checks
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 0
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 8: Invalid timeout
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 0
	cfg.Reporter.Type = "console"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 9: Invalid reporter type
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "invalid"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 10: Missing output file for JSON reporter
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "json"
	cfg.Reporter.OutputFile = ""

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))

	// Test case 11: Invalid schedule expression
	cfg = &Config{}
	cfg.AWS.Region = "us-west-2"
	cfg.Terraform.StateFile = "terraform.tfstate"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"
	cfg.App.ScheduleExpression = "invalid"

	err = cfg.Validate()
	assert.Error(t, err)
	assert.True(t, errors.IsValidationError(err))
}

func TestConfigGettersAndSetters(t *testing.T) {
	// Create a config
	cfg := &Config{}
	cfg.Detector.SourceOfTruth = "terraform"
	cfg.Detector.Attributes = []string{"instance_type", "ami"}
	cfg.Detector.ParallelChecks = 5
	cfg.Detector.TimeoutSeconds = 60
	cfg.Reporter.Type = "console"
	cfg.App.ScheduleExpression = "*/10 * * * *"

	// Test getters
	assert.Equal(t, "terraform", cfg.GetSourceOfTruth())
	assert.Equal(t, []string{"instance_type", "ami"}, cfg.GetAttributes())
	assert.Equal(t, 5, cfg.GetParallelChecks())
	assert.Equal(t, 60*time.Second, cfg.GetTimeout())
	assert.Equal(t, "console", cfg.GetReporterType())
	assert.Equal(t, "*/10 * * * *", cfg.GetScheduleExpression())

	// Test setters
	cfg.SetSourceOfTruth("aws")
	assert.Equal(t, "aws", cfg.GetSourceOfTruth())

	cfg.SetAttributes([]string{"instance_type", "ami", "tags"})
	assert.Equal(t, []string{"instance_type", "ami", "tags"}, cfg.GetAttributes())

	cfg.SetParallelChecks(10)
	assert.Equal(t, 10, cfg.GetParallelChecks())

	cfg.SetTimeout(120 * time.Second)
	assert.Equal(t, 120*time.Second, cfg.GetTimeout())
	assert.Equal(t, 120, cfg.Detector.TimeoutSeconds)

	cfg.SetReporterType("json")
	assert.Equal(t, "json", cfg.GetReporterType())

	cfg.SetScheduleExpression("0 */6 * * *")
	assert.Equal(t, "0 */6 * * *", cfg.GetScheduleExpression())
}
