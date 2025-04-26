package model

import (
	"time"

	"github.com/google/uuid"
)

// DriftResult represents the result of a drift detection operation
type DriftResult struct {
	// ID is a unique identifier for the drift detection result
	ID           string `json:"id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`

	// SourceType indicates which configuration is considered the source of truth
	SourceType ResourceOrigin `json:"source_type"`

	// Timestamp when the drift detection was performed
	Timestamp time.Time `json:"timestamp"`

	// HasDrift indicates whether any drift was detected
	HasDrift bool `json:"has_drift"`

	// DriftedAttributes contains information about all detected drifts
	DriftedAttributes map[string]AttributeDrift `json:"drifted_attributes,omitempty"`
}

// NewDriftResult creates a new drift detection result
func NewDriftResult(instanceID string, sourceType ResourceOrigin) *DriftResult {
	return &DriftResult{
		ID:                generateUUID(),
		ResourceID:        instanceID,
		ResourceType:      "aws_instance",
		SourceType:        sourceType,
		Timestamp:         time.Now(),
		DriftedAttributes: make(map[string]AttributeDrift),
	}
}

// AddDriftedAttribute adds a drifted attribute to the result
func (r *DriftResult) AddDriftedAttribute(path string, source, target interface{}) {
	r.DriftedAttributes[path] = AttributeDrift{
		Path:        path,
		SourceValue: source,
		TargetValue: target,
		Changed:     true,
	}
	r.HasDrift = true
}

// SetDriftedAttributes sets the complete map of drifted attributes
func (r *DriftResult) SetDriftedAttributes(drifts map[string]AttributeDrift) {
	r.DriftedAttributes = drifts
	r.HasDrift = len(drifts) > 0
}

// generateUUID generates a simple UUID for the drift result
func generateUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return id.String()
}
