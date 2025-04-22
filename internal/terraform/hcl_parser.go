package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// HCLParser parses Terraform HCL files
type HCLParser struct {
	BaseParser
	filePath string
}

// NewHCLParser creates a new HCL parser
func NewHCLParser(filePath string, log *logrus.Logger) *HCLParser {
	return &HCLParser{
		BaseParser: BaseParser{log: log},
		filePath:   filePath,
	}
}

// GetEC2Instances returns all EC2 instances defined in the Terraform HCL file
func (p *HCLParser) GetEC2Instances(ctx context.Context) ([]models.EC2Instance, error) {
	fileInfo, err := os.Stat(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	var files []string
	if fileInfo.IsDir() {
		// If it's a directory, get all .tf files
		entries, err := os.ReadDir(p.filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".tf" {
				files = append(files, filepath.Join(p.filePath, entry.Name()))
			}
		}
	} else {
		// If it's a file, just use that
		files = append(files, p.filePath)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Terraform files found")
	}

	// Parse the files
	parser := hclparse.NewParser()
	var instances []models.EC2Instance

	for _, file := range files {
		f, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			p.log.Warn(fmt.Sprintf("Failed to parse %s: %s", file, diags.Error()))
			continue
		}

		content, diags := f.Body.Content(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
			},
		})
		if diags.HasErrors() {
			p.log.Warn(fmt.Sprintf("Failed to decode %s: %s", file, diags.Error()))
			continue
		}

		// Find EC2 instances
		for _, block := range content.Blocks {
			if block.Type == "resource" && len(block.Labels) >= 2 && block.Labels[0] == "aws_instance" {
				name := block.Labels[1]
				instance, err := p.parseEC2Block(block, name)
				if err != nil {
					p.log.Warn(fmt.Sprintf("Failed to parse EC2 instance %s: %v", name, err))
					continue
				}
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// parseEC2Block extracts EC2 instance details from a Terraform block
func (p *HCLParser) parseEC2Block(block *hcl.Block, name string) (models.EC2Instance, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return models.EC2Instance{}, fmt.Errorf("failed to get attributes: %s", diags.Error())
	}

	// Initialize EC2 instance with the name as a fallback for instanceID
	instance := models.EC2Instance{
		ID: name, // Will be overridden if 'id' attribute exists
	}

	// Parse attributes
	for attrName, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			p.log.Warn(fmt.Sprintf("Failed to evaluate attribute %s: %s", attrName, diags.Error()))
			continue
		}

		switch attrName {
		case "id", "instance_id":
			if !val.IsNull() && val.Type() == cty.String {
				instance.ID = val.AsString()
			}
		case "instance_type":
			if !val.IsNull() && val.Type() == cty.String {
				instance.InstanceType = val.AsString()
			}
		case "ami":
			if !val.IsNull() && val.Type() == cty.String {
				instance.AMI = val.AsString()
			}
		case "subnet_id":
			if !val.IsNull() && val.Type() == cty.String {
				instance.SubnetID = val.AsString()
			}
		case "vpc_security_group_ids":
			if !val.IsNull() && val.Type().IsListType() {
				for it := val.ElementIterator(); it.Next(); {
					_, v := it.Element()
					if v.Type() == cty.String {
						instance.SecurityGroupIDs = append(instance.SecurityGroupIDs, v.AsString())
					}
				}
			}
		case "tags":
			if !val.IsNull() && val.Type().IsMapType() {
				instance.Tags = make(map[string]string)
				for it := val.ElementIterator(); it.Next(); {
					k, v := it.Element()
					if k.Type() == cty.String && v.Type() == cty.String {
						instance.Tags[k.AsString()] = v.AsString()
					}
				}
			}
		}
	}

	// If the instanceID is still just the name, add a note
	if instance.ID == name {
		p.log.Warn(fmt.Sprintf("No instance ID found for %s, using resource name as fallback", name))
	}

	return instance, nil
}
