package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/zclconf/go-cty/cty"
)

// HCLParser parses Terraform HCL configuration files
type HCLParser struct {
	logger *logging.Logger
}

// NewHCLParser creates a new Terraform HCL parser
func NewHCLParser(logger *logging.Logger) *HCLParser {
	return &HCLParser{
		logger: logger.WithField("component", "terraform-hcl"),
	}
}

// TerraformConfig represents the structure of Terraform configuration
type TerraformConfig struct {
	Resources []TerraformConfigResource `hcl:"resource,block"`
}

// TerraformConfigResource represents a resource block in Terraform configuration
type TerraformConfigResource struct {
	Type       string   `hcl:"type,label"`
	Name       string   `hcl:"name,label"`
	Attributes hcl.Body `hcl:",remain"`
}

// ParseHCLDir parses all .tf files in a directory
func (p *HCLParser) ParseHCLDir(ctx context.Context, dirPath string) ([]*model.Instance, error) {
	p.logger.Info(fmt.Sprintf("Parsing Terraform HCL files in directory: %s", dirPath))

	// Get all .tf files in the directory
	files, err := filepath.Glob(filepath.Join(dirPath, "*.tf"))
	if err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to list Terraform files in %s", dirPath), err)
	}

	p.logger.Info(fmt.Sprintf("Found %d Terraform files in directory", len(files)))
	if len(files) == 0 {
		return nil, errors.NewOperationalError(fmt.Sprintf("No Terraform files found in %s", dirPath), nil)
	}

	var instances []*model.Instance

	// Process each file
	for _, file := range files {
		fileInstances, err := p.ParseHCLFile(ctx, file)
		if err != nil {
			p.logger.Warn(fmt.Sprintf("Error parsing file %s: %v", file, err))
			continue
		}

		instances = append(instances, fileInstances...)
	}

	p.logger.Info(fmt.Sprintf("Found %d EC2 instances in Terraform configuration", len(instances)))
	return instances, nil
}

// ParseHCLFile parses a single Terraform HCL file
func (p *HCLParser) ParseHCLFile(ctx context.Context, filePath string) ([]*model.Instance, error) {
	p.logger.Info(fmt.Sprintf("Parsing Terraform HCL file: %s", filePath))

	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to read Terraform file: %s", filePath), err)
	}

	// Parse the HCL configuration
	var config TerraformConfig
	err = hclsimple.Decode(filePath, fileContent, nil, &config)
	if err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to decode HCL in %s", filePath), err)
	}

	var instances []*model.Instance

	// Process each resource
	for _, resource := range config.Resources {
		// Only process aws_instance resources
		if resource.Type == "aws_instance" {
			instance, err := p.extractInstanceFromResource(resource)
			if err != nil {
				p.logger.Warn(fmt.Sprintf("Failed to extract instance from resource %s: %v", resource.Name, err))
				continue
			}

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// extractInstanceFromResource extracts an EC2 instance from a Terraform resource
func (p *HCLParser) extractInstanceFromResource(resource TerraformConfigResource) (*model.Instance, error) {
	// Extract attributes from the resource
	attrs, err := p.extractAttributes(resource.Attributes)
	if err != nil {
		return nil, err
	}

	// Generate a pseudo-ID since the real ID won't be known until Terraform applies the configuration
	id := fmt.Sprintf("tf-%s-%s", resource.Type, resource.Name)

	// Add resource name and type to attributes
	attrs["resource_name"] = resource.Name
	attrs["resource_type"] = resource.Type

	return model.NewInstance(id, attrs, model.OriginTerraform), nil
}

// extractAttributes extracts attributes from HCL body
func (p *HCLParser) extractAttributes(body hcl.Body) (map[string]interface{}, error) {
	// Define a schema for the attributes
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "ami", Required: false},
			{Name: "instance_type", Required: false},
			{Name: "subnet_id", Required: false},
			{Name: "vpc_security_group_ids", Required: false},
			{Name: "key_name", Required: false},
			{Name: "availability_zone", Required: false},
			{Name: "tags", Required: false},
			{Name: "ebs_optimized", Required: false},
			{Name: "monitoring", Required: false},
			{Name: "iam_instance_profile", Required: false},
			{Name: "root_block_device", Required: false},
			{Name: "ebs_block_device", Required: false},
			{Name: "user_data", Required: false},
			{Name: "user_data_base64", Required: false},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "ebs_block_device"},
			{Type: "root_block_device"},
			{Type: "network_interface"},
			{Type: "timeouts"},
		},
	}

	// Extract content from body
	content, _, diags := body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, errors.NewOperationalError("Failed to extract attributes from HCL body", diags)
	}

	// Create evaluation context
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Extract attributes
	attrs := make(map[string]interface{})
	for name, attr := range content.Attributes {
		// Evaluate the expression
		value, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			p.logger.Warn(fmt.Sprintf("Failed to evaluate attribute %s: %v", name, diags.Error()))
			continue
		}

		// Convert the value to Go type
		attrs[name] = p.convertCtyValue(value)
	}

	// Process blocks (like ebs_block_device)
	for _, block := range content.Blocks {
		blockType := block.Type

		// Process the block content
		blockAttrs, err := p.extractAttributes(block.Body)
		if err != nil {
			p.logger.Warn(fmt.Sprintf("Failed to extract attributes from block %s: %v", blockType, err))
			continue
		}

		// Add the block to attributes
		if existing, ok := attrs[blockType]; ok {
			// If it's already a slice, append to it
			if slice, ok := existing.([]interface{}); ok {
				attrs[blockType] = append(slice, blockAttrs)
			} else {
				// Otherwise, create a new slice
				attrs[blockType] = []interface{}{blockAttrs}
			}
		} else {
			// First occurrence
			attrs[blockType] = []interface{}{blockAttrs}
		}
	}

	return attrs, nil
}

// convertCtyValue converts a cty.Value to a Go value
func (p *HCLParser) convertCtyValue(value cty.Value) interface{} {
	// Handle null values
	if value.IsNull() {
		return nil
	}

	// Handle different types
	switch {
	case value.Type() == cty.String:
		return value.AsString()
	case value.Type() == cty.Number:
		f, _ := value.AsBigFloat().Float64()
		return f
	case value.Type() == cty.Bool:
		return value.True()
	case value.Type().IsMapType():
		return p.convertCtyMap(value)
	case value.Type().IsListType() || value.Type().IsSetType():
		return p.convertCtyList(value)
	case value.Type().IsTupleType():
		return p.convertCtyTuple(value)
	case value.Type().IsObjectType():
		return p.convertCtyObject(value)
	default:
		return value.GoString()
	}
}

// convertCtyMap converts a cty map to a Go map
func (p *HCLParser) convertCtyMap(value cty.Value) map[string]interface{} {
	result := make(map[string]interface{})

	if value.IsNull() || !value.IsKnown() {
		return result
	}

	for k, v := range value.AsValueMap() {
		result[k] = p.convertCtyValue(v)
	}

	return result
}

// convertCtyList converts a cty list or set to a Go slice
func (p *HCLParser) convertCtyList(value cty.Value) []interface{} {
	result := make([]interface{}, 0)

	if value.IsNull() || !value.IsKnown() {
		return result
	}

	for _, v := range value.AsValueSlice() {
		result = append(result, p.convertCtyValue(v))
	}

	return result
}

// convertCtyTuple converts a cty tuple to a Go slice
func (p *HCLParser) convertCtyTuple(value cty.Value) []interface{} {
	return p.convertCtyList(value)
}

// convertCtyObject converts a cty object to a Go map
func (p *HCLParser) convertCtyObject(value cty.Value) map[string]interface{} {
	result := make(map[string]interface{})

	if value.IsNull() || !value.IsKnown() {
		return result
	}

	for k, v := range value.AsValueMap() {
		result[k] = p.convertCtyValue(v)
	}

	return result
}

// GetInstanceByName gets an EC2 instance by its resource name from HCL files
func (p *HCLParser) GetInstanceByName(ctx context.Context, dirPath, resrcName string) (*model.Instance, error) {
	// Get all instances
	instances, err := p.ParseHCLDir(ctx, dirPath)
	if err != nil {
		return nil, err
	}

	// Find the instance with the specified resource name
	for _, instance := range instances {
		if resourceName, ok := instance.Attributes["resource_name"].(string); ok && resourceName == resrcName {
			return instance, nil
		}
	}

	return nil, errors.NewNotFoundError("EC2 Instance Resource", resrcName)
}
