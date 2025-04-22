package terraform_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
)

func TestStateParser_GetEC2Instances_Success(t *testing.T) {
	log := logrus.New()
	stateFile := filepath.Join("testdata", "example.tfstate")

	parser := terraform.NewStateParser(stateFile, log)
	instances, err := parser.GetEC2Instances(context.Background())

	require.NoError(t, err)
	require.Len(t, instances, 1)

	ec2 := instances[0]
	require.Equal(t, "i-1234567890abcdef0", ec2.ID)
	require.Equal(t, "t2.micro", ec2.InstanceType)
	require.Equal(t, "ami-abc123", ec2.AMI)
	require.Equal(t, "subnet-xyz", ec2.SubnetID)
	require.ElementsMatch(t, []string{"sg-123", "sg-456"}, ec2.SecurityGroupIDs)
	require.Equal(t, "test-instance", ec2.Tags["Name"])
}

func TestStateParser_GetEC2Instances_FileNotFound(t *testing.T) {
	log := logrus.New()
	parser := terraform.NewStateParser("invalid.tfstate", log)

	instances, err := parser.GetEC2Instances(context.Background())
	require.Error(t, err)
	require.Nil(t, instances)
}
