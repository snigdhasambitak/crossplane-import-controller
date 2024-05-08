package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type VM struct {
	Name string `json:"name"`
}

func getVMNames(projectID string) ([]string, error) {
	fmt.Println("Fetching VM names from GCP...")
	cmd := exec.Command("gcloud", "compute", "instances", "list", "--project", projectID, "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var vms []VM
	err = json.Unmarshal(output, &vms)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, vm := range vms {
		names = append(names, vm.Name)
	}
	fmt.Println("Fetched VM names:", names)
	return names, nil
}

func renderCrossplaneTemplate(vmName string) string {
	return fmt.Sprintf(`
apiVersion: compute.gcp.upbound.io/v1beta1
kind: Instance
metadata:
  name: %s
  annotations:
    crossplane.io/external-name: %s
spec:
  managementPolicies: ["Observe"]
  forProvider:
    zone: europe-west1-c
`, vmName, vmName)
}

func applyCrossplaneConfig(template string, vmName string, errCh chan<- error, wg *sync.WaitGroup, doneCh chan<- struct{}) {
	defer wg.Done()

	// Generate unique filename based on VM name
	filename := vmName + ".yaml"

	f, err := os.Create(filename)
	if err != nil {
		errCh <- fmt.Errorf("error creating file: %w", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(template)
	if err != nil {
		errCh <- fmt.Errorf("error writing template to file: %w", err)
		return
	}

	cmd := exec.Command("kubectl", "apply", "-f", filename)
	fmt.Printf("Applying Crossplane config for VM %s...\n", vmName)
	err = cmd.Run()
	if err != nil {
		errCh <- fmt.Errorf("error applying config for VM %s: %w", vmName, err)
	} else {
		fmt.Printf("Crossplane config applied successfully for VM %s\n", vmName)
	}

	// Signal completion of this goroutine
	doneCh <- struct{}{}
}

func main() {
	projectID := "playground-common-cros1"

	vmNames, err := getVMNames(projectID)
	if err != nil {
		fmt.Println("Error fetching VM names:", err)
		return
	}

	errCh := make(chan error, len(vmNames))
	doneCh := make(chan struct{}, len(vmNames))

	// Fetch VM names first and store them in a slice
	var vmTemplates []string
	for _, vmName := range vmNames {
		vmTemplates = append(vmTemplates, renderCrossplaneTemplate(vmName))
	}

	// Apply Crossplane configs in parallel
	var wg sync.WaitGroup
	for i, template := range vmTemplates {
		wg.Add(1)
		go applyCrossplaneConfig(template, vmNames[i], errCh, &wg, doneCh)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	// Wait for all done signals
	for range vmNames {
		<-doneCh
	}

	// Close the error channel after all goroutines have finished
	close(errCh)

	// Print any errors encountered during Crossplane config application
	for err := range errCh {
		fmt.Println(err)
	}
}
