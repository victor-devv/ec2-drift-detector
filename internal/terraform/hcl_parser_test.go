package terraform_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
)

func TestHCLParser_GetEC2Instances_Success(t *testing.T) {
	log := logrus.New()
	tfFile := filepath.Join("testdata", "main.tf")

	parser := terraform.NewHCLParser(tfFile, log)
	instances, err := parser.GetEC2Instances(context.Background())

	require.NoError(t, err)
	require.Len(t, instances, 1)

	ec2 := instances[0]
	require.Equal(t, "i-abc123", ec2.ID)
	require.Equal(t, "ami-abc123", ec2.AMI)
	require.Equal(t, "t3.medium", ec2.InstanceType)
	require.Equal(t, "subnet-xyz", ec2.SubnetID)
	require.ElementsMatch(t, []string{"sg-123"}, ec2.SecurityGroupIDs)
	require.Equal(t, "production", ec2.Tags["env"])
}

func TestHCLParser_GetEC2Instances_InvalidPath(t *testing.T) {
	log := logrus.New()
	parser := terraform.NewHCLParser("nonexistent.tf", log)

	instances, err := parser.GetEC2Instances(context.Background())
	require.Error(t, err)
	require.Nil(t, instances)
}
