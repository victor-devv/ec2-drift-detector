package detector_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsinternal "github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/detector"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

type mockEC2API struct {
	instances []types.Instance
}

func (m *mockEC2API) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: m.instances,
			},
		},
	}, nil
}

type mockTFParser struct {
	instances []models.EC2Instance
}

func (m *mockTFParser) GetEC2Instances(_ context.Context) ([]models.EC2Instance, error) {
	return m.instances, nil
}

func TestDetectDrift_InstanceTypeMismatch(t *testing.T) {
	log := logrus.New()
	mockAWS := &mockEC2API{
		instances: []types.Instance{
			{
				InstanceId:   aws.String("i-abc123"),
				InstanceType: types.InstanceTypeT2Micro,
				ImageId:      aws.String("ami-001"),
			},
		},
	}
	mockTF := &mockTFParser{
		instances: []models.EC2Instance{
			{
				ID:           "i-abc123",
				InstanceType: "t3.micro",
				AMI:          "ami-001",
			},
		},
	}

	ec2Client := awsinternal.NewTestEC2Client(mockAWS, log)
	det := detector.NewEC2Detector(*ec2Client, mockTF, log)

	results, err := det.DetectDrift(context.Background(), []string{"instance_type"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].Drifted)
	require.Equal(t, "instance_type", results[0].DriftDetails[0].Attribute)
}

func TestDetectDrift_TagMismatch(t *testing.T) {
	log := logrus.New()
	mockAWS := &mockEC2API{
		instances: []types.Instance{
			{
				InstanceId:   aws.String("i-xyz456"),
				InstanceType: types.InstanceTypeT2Micro,
				ImageId:      aws.String("ami-001"),
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String("prod")},
				},
			},
		},
	}
	mockTF := &mockTFParser{
		instances: []models.EC2Instance{
			{
				ID:           "i-xyz456",
				InstanceType: "t2.micro",
				AMI:          "ami-001",
				Tags: map[string]string{
					"Name": "dev",
				},
			},
		},
	}

	ec2Client := awsinternal.NewTestEC2Client(mockAWS, log)
	det := detector.NewEC2Detector(*ec2Client, mockTF, log)

	results, err := det.DetectDrift(context.Background(), []string{"tags"})
	require.NoError(t, err)
	require.True(t, results[0].Drifted)
	require.Equal(t, "tags.Name", results[0].DriftDetails[0].Attribute)
	require.Equal(t, "prod", results[0].DriftDetails[0].AWSValue)
	require.Equal(t, "dev", results[0].DriftDetails[0].TerraformValue)
}

func TestDetectDrift_MissingInTerraform(t *testing.T) {
	log := logrus.New()
	awsOnlyInstanceID := "i-not-in-tf"

	mockAWS := &mockEC2API{
		instances: []types.Instance{
			{
				InstanceId:   aws.String(awsOnlyInstanceID),
				InstanceType: types.InstanceTypeT2Micro,
				ImageId:      aws.String("ami-002"),
			},
		},
	}
	mockTF := &mockTFParser{
		instances: []models.EC2Instance{
			// Include an unrelated TF ID to simulate that AWS-only instance is unexpected
			{ID: "i-different-id"},
		},
	}

	ec2Client := awsinternal.NewTestEC2Client(mockAWS, log)
	det := detector.NewEC2Detector(*ec2Client, mockTF, log)

	results, err := det.DetectDrift(context.Background(), []string{"instance_type"})
	require.NoError(t, err)

	// Should find the AWS-only instance as drifted
	var found bool
	for _, res := range results {
		if res.ResourceID == awsOnlyInstanceID && res.Drifted {
			found = true
			require.Equal(t, "existence", res.DriftDetails[0].Attribute)
			require.Equal(t, "exists", res.DriftDetails[0].AWSValue)
			require.Equal(t, "not_exists", res.DriftDetails[0].TerraformValue)
		}
	}
	require.True(t, found, "Expected AWS-only instance to be marked as drifted")
}

func TestDetectDrift_NoDrift(t *testing.T) {
	log := logrus.New()
	mockAWS := &mockEC2API{
		instances: []types.Instance{
			{
				InstanceId:   aws.String("i-consistent"),
				InstanceType: types.InstanceTypeT2Micro,
				ImageId:      aws.String("ami-123"),
			},
		},
	}
	mockTF := &mockTFParser{
		instances: []models.EC2Instance{
			{
				ID:           "i-consistent",
				InstanceType: "t2.micro",
				AMI:          "ami-123",
			},
		},
	}

	ec2Client := awsinternal.NewTestEC2Client(mockAWS, log)
	det := detector.NewEC2Detector(*ec2Client, mockTF, log)

	results, err := det.DetectDrift(context.Background(), []string{"instance_type", "ami"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.False(t, results[0].Drifted)
}
