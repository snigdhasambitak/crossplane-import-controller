package cloudprovider

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// GCPProvider struct represents the GCP cloud provider
type GCPProvider struct{}

// NewGCPProvider creates a new instance of the GCP cloud provider
func NewGCPProvider() *GCPProvider {
	return &GCPProvider{}
}

// GetVMNames retrieves the names of VMs from GCP
func (gcp *GCPProvider) GetVMNames(projectID string) ([]string, error) {
	cmd := exec.Command("gcloud", "compute", "instances", "list", "--project", projectID, "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error fetching VM names: %w", err)
	}

	var vms []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &vms); err != nil {
		return nil, fmt.Errorf("error unmarshalling VM names: %w", err)
	}

	var names []string
	for _, vm := range vms {
		names = append(names, vm.Name)
	}
	return names, nil
}
