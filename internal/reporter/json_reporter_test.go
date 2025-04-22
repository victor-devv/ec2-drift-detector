package reporter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
	"github.com/victor-devv/ec2-drift-detector/internal/reporter"
)

func TestJSONReporter_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewJSONReporter(log).WithWriter(&buf)

	results := []models.DriftResult{
		{
			ResourceID:   "i-abc",
			ResourceType: "aws_instance",
			InTerraform:  true,
			InAWS:        true,
			Drifted:      true,
			DriftDetails: []models.AttributeDiff{
				{
					Attribute:      "ami",
					AWSValue:       "ami-123",
					TerraformValue: "ami-456",
				},
			},
		},
	}

	err := r.Report(context.Background(), results)
	require.NoError(t, err)

	var parsed reporter.JSONReport
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	require.Equal(t, 1, parsed.Summary.TotalResources)
	require.Equal(t, 1, parsed.Summary.DriftedResources)
	require.Equal(t, 0, parsed.Summary.NonDriftedResources)
	require.Equal(t, "i-abc", parsed.Results[0].ResourceID)
	require.Equal(t, "ami", parsed.Results[0].DriftDetails[0].Attribute)
}

func TestJSONReporter_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewJSONReporter(log).WithWriter(&buf)

	results := []models.DriftResult{
		{
			ResourceID:   "i-xyz",
			ResourceType: "aws_instance",
			InTerraform:  true,
			InAWS:        true,
			Drifted:      false,
		},
	}

	err := r.Report(context.Background(), results)
	require.NoError(t, err)

	var parsed reporter.JSONReport
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	require.Equal(t, 1, parsed.Summary.TotalResources)
	require.Equal(t, 0, parsed.Summary.DriftedResources)
	require.Equal(t, 1, parsed.Summary.NonDriftedResources)
}

func TestJSONReporter_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()

	r := reporter.NewJSONReporter(log).WithWriter(&buf)

	err := r.Report(context.Background(), nil)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, `"totalResources":0`)
	require.Contains(t, output, `"results":[]`)
}
