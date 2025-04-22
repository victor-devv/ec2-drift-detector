package reporter_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
	"github.com/victor-devv/ec2-drift-detector/internal/reporter"
)

func TestConsoleReporter_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewConsoleReporter(log).WithWriter(&buf)

	results := []models.DriftResult{
		{
			ResourceID:   "i-abc",
			ResourceType: "aws_instance",
			InTerraform:  true,
			InAWS:        true,
			Drifted:      false,
		},
	}

	err := r.Report(context.Background(), results)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "✓ No drift detected")
	require.Contains(t, output, "i-abc")
	require.Contains(t, output, "OK")
}

func TestConsoleReporter_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewConsoleReporter(log).WithWriter(&buf)

	results := []models.DriftResult{
		{
			ResourceID:   "i-abc",
			ResourceType: "aws_instance",
			InTerraform:  true,
			InAWS:        true,
			Drifted:      true,
			DriftDetails: []models.AttributeDiff{
				{
					Attribute:      "instance_type",
					AWSValue:       "t2.micro",
					TerraformValue: "t3.micro",
				},
			},
		},
	}

	err := r.Report(context.Background(), results)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "✗ Drift detected")
	require.Contains(t, output, "DRIFTED")
	require.Contains(t, output, "instance_type: AWS='t2.micro', TF='t3.micro'")
}

func TestConsoleReporter_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewConsoleReporter(log).WithWriter(&buf)

	err := r.Report(context.Background(), []models.DriftResult{})
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "No resources found")
}
