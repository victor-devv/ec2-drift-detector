package models

// EC2Instance represents the common structure for EC2 instances from both AWS and Terraform
type EC2Instance struct {
	ID                    string            `json:"id"`
	InstanceType          string            `json:"instance_type"`
	AMI                   string            `json:"ami"`
	SubnetID              string            `json:"subnet_id"`
	VPCID                 string            `json:"vpc_id"`
	SecurityGroupIDs      []string          `json:"security_group_ids"`
	SecurityGroupNames    []string          `json:"security_group_names"`
	KeyName               string            `json:"key_name"`
	IAMRole               string            `json:"iam_role"`
	PublicIP              string            `json:"public_ip"`
	PrivateIP             string            `json:"private_ip"`
	Tags                  map[string]string `json:"tags"`
	RootVolumeSize        int64             `json:"root_volume_size"`
	RootVolumeType        string            `json:"root_volume_type"`
	UserData              string            `json:"user_data"`
	EBSOptimized          bool              `json:"ebs_optimized"`
	SourceDestCheck       bool              `json:"source_dest_check"`
	MonitoringEnabled     bool              `json:"monitoring_enabled"`
	TerminationProtection bool              `json:"termination_protection"`
}

// DriftResult represents the differences between AWS and Terraform EC2 resources
type DriftResult struct {
	ResourceID   string          `json:"resourceId"`
	ResourceType string          `json:"resourceType"`
	InTerraform  bool            `json:"inTerraform"`
	InAWS        bool            `json:"inAWS"`
	Drifted      bool            `json:"drifted"`
	DriftDetails []AttributeDiff `json:"attribute_diffs"`
}

//map[string]AttributeDiff

// AttributeDiff represents the difference for a single attribute
type AttributeDiff struct {
	Attribute      string      `json:"attribute_name"`
	AWSValue       interface{} `json:"aws_value"`
	TerraformValue interface{} `json:"terraform_value"`
	IsComplex      bool        `json:"is_complex"`
}
