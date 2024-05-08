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
	"time"
)

const (
	configFolder           = "config"
	instanceTemplateFolder = "instanceTemplates"
	knownVMsFile           = "known_vms.json"
)

type KnownVMs struct {
	VMs map[string]struct{}
}

func saveKnownVMs(knownVMs KnownVMs) error {
	data, err := json.Marshal(knownVMs)
	if err != nil {
		return fmt.Errorf("error marshalling known VMs: %w", err)
	}
	err = ioutil.WriteFile(knownVMsFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing known VMs to file: %w", err)
	}
	return nil
}

func loadKnownVMs() (KnownVMs, error) {
	data, err := ioutil.ReadFile(knownVMsFile)
	if err != nil {
		return KnownVMs{}, fmt.Errorf("error reading known VMs file: %w", err)
	}
	var knownVMs KnownVMs
	err = json.Unmarshal(data, &knownVMs)
	if err != nil {
		return KnownVMs{}, fmt.Errorf("error unmarshalling known VMs: %w", err)
	}
	return knownVMs, nil
}

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

func applyCrossplaneConfig(vmName string, errCh chan<- error, wg *sync.WaitGroup, doneCh chan<- struct{}) {
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

	// Load known VMs from file
	knownVMs, err := loadKnownVMs()
	if err != nil {
		log.Println("Error loading known VMs:", err)
		knownVMs = KnownVMs{VMs: make(map[string]struct{})}
	}

	// Main loop to continuously monitor for new VMs
	for {
		// Fetch VM names from GCP
		vmNames, err := getVMNames(projectID)
		if err != nil {
			log.Println("Error fetching VM names:", err)
			continue
		}

		// Check for new VMs
		for _, vmName := range vmNames {
			if _, ok := knownVMs.VMs[vmName]; !ok {
				knownVMs.VMs[vmName] = struct{}{}
				// Save known VMs to file
				err := saveKnownVMs(knownVMs)
				if err != nil {
					log.Println("Error saving known VMs:", err)
				}
				// Apply Crossplane configuration for new VM
				var wg sync.WaitGroup
				errCh := make(chan error)
				doneCh := make(chan struct{})

				wg.Add(1)
				go applyCrossplaneConfig(vmName, errCh, &wg, doneCh)

				go func() {
					wg.Wait()
					close(doneCh)
				}()

				go func() {
					for err := range errCh {
						log.Println("Error applying Crossplane config:", err)
					}
				}()

				<-doneCh
				close(errCh)
			}
		}

		// Check for deleted VMs
		for knownVM := range knownVMs.VMs {
			found := false
			for _, vmName := range vmNames {
				if knownVM == vmName {
					found = true
					break
				}
			}
			if !found {
				// Delete the Crossplane configuration for the deleted VM
				delete(knownVMs.VMs, knownVM)
				instanceTemplateFilename := fmt.Sprintf("%s/%s.yaml", instanceTemplateFolder, knownVM)
				cmd := exec.Command("kubectl", "delete", "-f", instanceTemplateFilename)
				log.Printf("Deleting Crossplane config for VM %s...\n", knownVM)
				err := cmd.Run()
				if err != nil {
					log.Printf("Error deleting Crossplane config for VM %s: %v\n", knownVM, err)
				} else {
					log.Printf("Crossplane config for VM %s deleted successfully\n", knownVM)
					// Save known VMs to file
					err := saveKnownVMs(knownVMs)
					if err != nil {
						log.Println("Error saving known VMs:", err)
					}
				}
			}
		}

		// Sleep for a while before checking for new VMs again
		time.Sleep(2 * time.Minute) // Adjust the interval as needed
	}
}
