package model

// TFState represents the structure of a Terraform state file
type TFState struct {
	Version          int                    `json:"version"`
	TerraformVersion string                 `json:"terraform_version"`
	Serial           int                    `json:"serial"`
	Lineage          string                 `json:"lineage"`
	Resources        []TFResource           `json:"resources"`
	Outputs          map[string]interface{} `json:"outputs"`
}

// TFResource represents a resource in a Terraform state file
type TFResource struct {
	Module    string                 `json:"module"`
	Mode      string                 `json:"mode"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Provider  string                 `json:"provider"`
	Instances []TFResourceInstance   `json:"instances"`
	Values    map[string]interface{} `json:"values"`
}

// TFResourceInstance represents an instance of a resource in a Terraform state file
type TFResourceInstance struct {
	IndexKey   interface{}            `json:"index_key"`
	Status     string                 `json:"status"`
	Schema     map[string]interface{} `json:"schema_version"`
	Attributes map[string]interface{} `json:"attributes"`
	Private    string                 `json:"private"`
}
