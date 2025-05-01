package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

func TestTFState_UnmarshalBasic(t *testing.T) {
	input := `{
		"version": 4,
		"terraform_version": "1.4.0",
		"serial": 1,
		"lineage": "abc-123",
		"outputs": {
			"example_output": "value"
		},
		"resources": [
			{
				"module": "module.example",
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"index_key": 0,
						"status": "tainted",
						"schema_version": 1,
						"attributes": {
							"id": "i-1234567890",
							"instance_type": "t2.micro"
						},
						"private": "xxxx"
					}
				]
			}
		]
	}`

	var state model.TFState
	err := json.Unmarshal([]byte(input), &state)
	assert.NoError(t, err)
	assert.Equal(t, 4, state.Version)
	assert.Equal(t, "1.4.0", state.TerraformVersion)
	assert.Equal(t, "abc-123", state.Lineage)
	assert.Equal(t, 1, state.Serial)
	assert.Contains(t, state.Outputs, "example_output")
	assert.Len(t, state.Resources, 1)

	r := state.Resources[0]
	assert.Equal(t, "aws_instance", r.Type)
	assert.Equal(t, "web", r.Name)
	assert.Equal(t, "module.example", r.Module)
	assert.Len(t, r.Instances, 1)

	inst := r.Instances[0]
	assert.Equal(t, "tainted", inst.Status)
	assert.Equal(t, 1, inst.Schema)
	assert.Equal(t, "i-1234567890", inst.Attributes["id"])
	assert.Equal(t, "t2.micro", inst.Attributes["instance_type"])
	assert.Equal(t, "xxxx", inst.Private)
}
