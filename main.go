package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func getVMNames(projectID string) ([]string, error) {
	log.Println("Fetching VM names from GCP...")
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
	log.Println("Fetched VM names:", names)
	return names, nil
}

func readConfigFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", path, err)
	}
	return string(content), nil
}

func writeConfigFile(content, filename string) error {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing config to file %s: %w", filename, err)
	}
	return nil
}

func applyCrossplaneConfig(vmName, configFolder, instanceTemplateFolder string, errCh chan<- error, wg *sync.WaitGroup, doneCh chan<- struct{}) {
	defer wg.Done()

	// Step 1: Read the instance template from instance_template.yaml
	template, err := readConfigFile(fmt.Sprintf("%s/instance_template.yaml", configFolder))
	if err != nil {
		errCh <- err
		return
	}

	// Step 2: Replace placeholders with VM name
	instanceTemplate := strings.ReplaceAll(template, "<vmName>", vmName)

	// Step 3: Write the rendered instance template to a file in the instanceTemplate folder
	instanceTemplateFilename := fmt.Sprintf("%s/%s.yaml", instanceTemplateFolder, vmName)
	err = writeConfigFile(instanceTemplate, instanceTemplateFilename)
	if err != nil {
		errCh <- err
		return
	}

	// Step 4: Apply the Crossplane configuration using kubectl apply
	cmd := exec.Command("kubectl", "apply", "-f", instanceTemplateFilename)
	log.Printf("Applying Crossplane config for VM %s...\n", vmName)
	err = cmd.Run()
	if err != nil {
		errCh <- fmt.Errorf("error applying config for VM %s: %w", vmName, err)
		return // Exit function if applying config fails
	}

	// Step 5: Signal that the processing for this VM is done
	doneCh <- struct{}{}
}

func main() {
	// Retrieve project ID from environment variable
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is not set")
	}

	configFolder := "config"
	instanceTemplateFolder := "instanceTemplates"

	// Fetch VM names from GCP
	vmNames, err := getVMNames(projectID)
	if err != nil {
		log.Fatal("Error fetching VM names:", err)
	}

	errCh := make(chan error, len(vmNames))
	doneCh := make(chan struct{}, len(vmNames))

	var wg sync.WaitGroup
	for _, vmName := range vmNames {
		wg.Add(1)
		go applyCrossplaneConfig(vmName, configFolder, instanceTemplateFolder, errCh, &wg, doneCh)
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for range vmNames {
		<-doneCh
	}

	close(errCh)

	for err := range errCh {
		log.Println(err)
	}
}
