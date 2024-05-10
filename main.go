package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/snigdhasambitak/crossplane-import-controller/cloudprovider"
	"github.com/snigdhasambitak/crossplane-import-controller/config"
	"github.com/snigdhasambitak/crossplane-import-controller/template"
)

const (
// Define constants here
)

func main() {
	// Retrieve project ID from environment variable
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is not set")
	}

	// Load known VMs from file
	knownVMs, err := config.LoadKnownVMs()
	if err != nil {
		log.Println("Error loading known VMs:", err)
		knownVMs = config.NewKnownVMs()
	}

	// Create a new instance of the GCP cloud provider
	gcp := cloudprovider.NewGCPProvider()

	// Main loop to continuously monitor for new VMs
	for {
		// Fetch VM names from GCP
		vmNames, err := gcp.GetVMNames(projectID)
		if err != nil {
			log.Println("Error fetching VM names:", err)
			continue
		}

		// Check for new VMs
		for _, vmName := range vmNames {
			if knownVMs.Add(vmName) {
				// Apply Crossplane configuration for new VM
				var wg sync.WaitGroup
				errCh := make(chan error)
				doneCh := make(chan struct{})

				wg.Add(1)
				go template.ApplyCrossplaneConfig(vmName, errCh, &wg, doneCh)

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
		for knownVM := range knownVMs.Get() {
			found := false
			for _, vmName := range vmNames {
				if knownVM == vmName {
					found = true
					break
				}
			}
			if !found {
				// Delete the Crossplane configuration for the deleted VM
				knownVMs.Remove(knownVM)
				if err := template.DeleteCrossplaneConfig(knownVM); err != nil { // Change here
					log.Printf("Error deleting Crossplane config for VM %s: %v\n", knownVM, err)
				} else {
					log.Printf("Crossplane config for VM %s deleted successfully\n", knownVM)
					// Save known VMs to file
					if err := knownVMs.Save(); err != nil {
						log.Println("Error saving known VMs:", err)
					}
				}
			}
		}
		// Sleep for a while before checking for new VMs again
		time.Sleep(2 * time.Minute) // Adjust the interval as needed
	}
}
