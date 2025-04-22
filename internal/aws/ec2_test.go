package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	awsinternal "github.com/victor-devv/ec2-drift-detector/internal/aws"
)

// MockEC2API implements aws.EC2API
type MockEC2API struct {
	Response *ec2.DescribeInstancesOutput
	Err      error
}

func (m *MockEC2API) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.Response, m.Err
}

// ensure that AWS response is correctly parsed into models.EC2Instance, Tags, SubnetID, and SGs are handled AND Errors are wrapped and returned as expected
func TestDescribeInstances_Success(t *testing.T) {
	mockClient := &MockEC2API{
		Response: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:   aws.String("i-1234567890abcdef0"),
							ImageId:      aws.String("ami-abc123"),
							InstanceType: types.InstanceTypeT2Micro,
							SubnetId:     aws.String("subnet-xyz"),
							SecurityGroups: []types.GroupIdentifier{
								{
									GroupId: aws.String("sg-001"),
								},
							},
							Tags: []types.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("test-instance"),
								},
							},
						},
					},
				},
			},
		},
	}

	logger := logrus.New()
	impl := awsinternal.NewTestEC2Client(mockClient, logger)

	results, err := impl.DescribeInstances(context.Background(), []string{"i-1234567890abcdef0"})

	require.NoError(t, err)
	require.Len(t, results, 1)

	ec2Instance := results[0]
	require.Equal(t, "i-1234567890abcdef0", ec2Instance.ID)
	require.Equal(t, "ami-abc123", ec2Instance.AMI)
	require.Equal(t, "t2.micro", ec2Instance.InstanceType)
	require.Equal(t, "subnet-xyz", ec2Instance.SubnetID)
	require.ElementsMatch(t, []string{"sg-001"}, ec2Instance.SecurityGroupIDs)
	require.Equal(t, map[string]string{"Name": "test-instance"}, ec2Instance.Tags)
}
