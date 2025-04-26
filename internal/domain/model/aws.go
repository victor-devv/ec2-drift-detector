package model

import (
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/victor-devv/ec2-drift-detector/pkg/comparator"
)

// ResourceOrigin represents the source of a resource configuration
type ResourceOrigin string

const (
	// OriginAWS indicates the resource configuration comes from AWS
	OriginAWS ResourceOrigin = "aws"
	// OriginTerraform indicates the resource configuration comes from Terraform
	OriginTerraform ResourceOrigin = "terraform"
)

// Instance represents an EC2 instance configuration with attributes
type Instance struct {
	ID           string                 `json:"id"`
	InstanceType string                 `json:"instance_type"`
	Attributes   map[string]interface{} `json:"attributes"`
	Origin       ResourceOrigin         `json:"origin"`
}

// NewInstance creates a new instance with the given ID and attributes
func NewInstance(id string, attrs map[string]interface{}, origin ResourceOrigin) *Instance {
	instance := &Instance{
		ID:         id,
		Attributes: make(map[string]interface{}),
		Origin:     origin,
	}

	// Extract instance type from attributes if present
	if instType, ok := attrs["instance_type"].(string); ok {
		instance.InstanceType = instType
	}

	// Copy all attributes
	for k, v := range attrs {
		instance.Attributes[k] = v
	}

	return instance
}

// GetAttribute returns an attribute value by path using dot notation (e.g., "ebs_block_device.volume_size")
func (i *Instance) GetAttribute(path string) (interface{}, bool) {
	if path == "instance_type" {
		return i.InstanceType, true
	}

	return GetNestedValue(i.Attributes, path)
}

// GetNestedValue retrieves a value from a nested map structure using dot notation
func GetNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")

	var current interface{} = data

	for _, part := range parts {
		// Handle array indexing if needed
		// For simplicity, this implementation doesn't handle array indices

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

// CompareAttributes compares attributes between two instances using specified paths
// Returns a map of drifted attributes with both values
func CompareAttributes(source, target *Instance, attributePaths []string) map[string]AttributeDrift {
	result := make(map[string]AttributeDrift)
	var wg sync.WaitGroup
	resultMutex := sync.Mutex{}

	for _, path := range attributePaths {
		wg.Add(1)
		go func(attrPath string) {
			defer wg.Done()

			sourceVal, sourceExists := source.GetAttribute(attrPath)
			targetVal, targetExists := target.GetAttribute(attrPath)

			// Check for existence in both sources
			if !sourceExists && !targetExists {
				return
			}

			if !sourceExists || !targetExists {
				resultMutex.Lock()
				result[attrPath] = AttributeDrift{
					Path:        attrPath,
					SourceValue: sourceVal,
					TargetValue: targetVal,
					Changed:     true,
				}
				resultMutex.Unlock()
				return
			}

			// If both values exist, compare them
			if !reflect.DeepEqual(sourceVal, targetVal) {
				if attrPath == "tags" {
					comp := comparator.NewComparator()
					tagDrifts := comp.CompareDeep(sourceVal, targetVal)
					if len(tagDrifts) > 0 {
						resultMutex.Lock()
						result[attrPath] = AttributeDrift{
							Path:        attrPath,
							SourceValue: sourceVal,
							TargetValue: targetVal,
							Changed:     true,
						}
						resultMutex.Unlock()
					}
				} else {
					resultMutex.Lock()
					result[attrPath] = AttributeDrift{
						Path:        attrPath,
						SourceValue: sourceVal,
						TargetValue: targetVal,
						Changed:     true,
					}
					resultMutex.Unlock()
				}

			}
		}(path)
	}

	wg.Wait()
	return result
}

// AttributeDrift represents a detected drift for a specific attribute
type AttributeDrift struct {
	Path        string      `json:"path"`
	SourceValue interface{} `json:"source_value"`
	TargetValue interface{} `json:"target_value"`
	Changed     bool        `json:"changed"`
}

// NestedCompare implements deep comparison of nested attributes using goroutines
func NestedCompare(source, target map[string]interface{}, basePath string, maxDepth int, result *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()

	if maxDepth <= 0 {
		return
	}

	for key, sourceVal := range source {
		path := key
		if basePath != "" {
			path = basePath + "." + key
		}

		targetVal, exists := target[key]
		if !exists {
			result.Store(path, AttributeDrift{
				Path:        path,
				SourceValue: sourceVal,
				TargetValue: nil,
				Changed:     true,
			})
			continue
		}

		// If both values are maps, compare them recursively with goroutines
		sourceMap, sourceIsMap := sourceVal.(map[string]interface{})
		targetMap, targetIsMap := targetVal.(map[string]interface{})

		if sourceIsMap && targetIsMap {
			wg.Add(1)
			go NestedCompare(sourceMap, targetMap, path, maxDepth-1, result, wg)
		} else if !reflect.DeepEqual(sourceVal, targetVal) {
			result.Store(path, AttributeDrift{
				Path:        path,
				SourceValue: sourceVal,
				TargetValue: targetVal,
				Changed:     true,
			})
		}
	}

	// Check for keys in target that don't exist in source
	for key, targetVal := range target {
		path := key
		if basePath != "" {
			path = basePath + "." + key
		}

		if _, exists := source[key]; !exists {
			result.Store(path, AttributeDrift{
				Path:        path,
				SourceValue: nil,
				TargetValue: targetVal,
				Changed:     true,
			})
		}
	}
}

//======================================

// EC2Instance represents the AWS EC2 instance configuration
type EC2Instance struct {
	ID                    string                 `json:"id"`
	InstanceType          string                 `json:"instance_type"`
	AMI                   string                 `json:"ami"`
	Architecture          string                 `json:"architecture"`
	SubnetID              string                 `json:"subnet_id"`
	VPCID                 string                 `json:"vpc_id"`
	State                 string                 `json:"state"`
	LaunchTime            *time.Time             `json:"launch_time"`
	SecurityGroups        []SecurityGroup        `json:"security_groups"`
	SecurityGroupIDs      []string               `json:"security_group_ids"`
	SecurityGroupNames    []string               `json:"security_group_names"`
	EBSVolumes            []EBSVolume            `json:"ebs_volumes"`
	KeyName               string                 `json:"key_name"`
	IAMRole               string                 `json:"iam_role"`
	PublicIPAddress       string                 `json:"public_ip"`
	PublicDNSName         string                 `json:"public_dns_name"`
	PrivateIPAddress      string                 `json:"private_ip"`
	PrivateDNSName        string                 `json:"private_dns_name"`
	Tags                  map[string]string      `json:"tags"`
	RootVolumeSize        int64                  `json:"root_volume_size"`
	RootDeviceType        string                 `json:"root_volume_type"`
	UserData              string                 `json:"user_data"`
	EBSOptimized          bool                   `json:"ebs_optimized"`
	SourceDestCheck       bool                   `json:"source_dest_check"`
	MonitoringEnabled     bool                   `json:"monitoring_enabled"`
	TerminationProtection bool                   `json:"termination_protection"`
	Metadata              map[string]interface{} `json:"metadata"` // For additional attributes
}

// SecurityGroup represents an AWS security group
type SecurityGroup struct {
	GroupID   string
	GroupName string
}

// EBSVolume represents an EBS volume attached to an EC2 instance
type EBSVolume struct {
	VolumeID   string
	Size       int
	VolumeType string
	IOPS       int
	Encrypted  bool
	Tags       map[string]string
}

// FromAWSInstance converts AWS SDK EC2 instance to our domain model
func FromAWSInstance(instance types.Instance) EC2Instance {
	sgs := make([]SecurityGroup, 0, len(instance.SecurityGroups))
	for _, sg := range instance.SecurityGroups {
		sgs = append(sgs, SecurityGroup{
			GroupID:   *sg.GroupId,
			GroupName: *sg.GroupName,
		})
	}

	tags := make(map[string]string)
	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}

	var launchTime *time.Time
	if instance.LaunchTime != nil {
		launchTime = instance.LaunchTime
	}

	return EC2Instance{
		ID:               *instance.InstanceId,
		InstanceType:     string(instance.InstanceType),
		AMI:              *instance.ImageId,
		VPCID:            *instance.VpcId,
		SubnetID:         *instance.SubnetId,
		SecurityGroups:   sgs,
		Tags:             tags,
		State:            string(instance.State.Name),
		LaunchTime:       launchTime,
		PrivateDNSName:   *instance.PrivateDnsName,
		PrivateIPAddress: *instance.PrivateIpAddress,
		PublicDNSName:    *instance.PublicDnsName,
		PublicIPAddress:  *instance.PublicIpAddress,
		Architecture:     string(instance.Architecture),
		RootDeviceType:   string(instance.RootDeviceType),
	}
}

// Identifier returns the unique resource ID
func (e EC2Instance) Identifier() string {
	return e.ID
}
