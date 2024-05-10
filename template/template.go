package template

import (
	"fmt"
	"log"
	"sync"

	"github.com/snigdhasambitak/crossplane-import-controller/utils"
)

const (
	configFolder           = "resourcetemplates"
	instanceTemplateFolder = "generatedresources"
)

// ApplyCrossplaneConfig applies Crossplane configuration for the given VM
func ApplyCrossplaneConfig(vmName string, errCh chan<- error, wg *sync.WaitGroup, doneCh chan<- struct{}) {
	defer wg.Done()

	// Step 1: Read the instance template from instance_template.yaml
	template, err := utils.ReadConfigFile(fmt.Sprintf("%s/instance_template.yaml", configFolder))
	if err != nil {
		errCh <- err
		return
	}

	// Step 2: Replace placeholders with VM name
	instanceTemplate := utils.ReplacePlaceholder(template, "<vmName>", vmName)

	// Step 3: Write the rendered instance template to a file in the instanceTemplate folder
	instanceTemplateFilename := fmt.Sprintf("%s/%s.yaml", instanceTemplateFolder, vmName)
	if err := utils.WriteConfigFile(instanceTemplate, instanceTemplateFilename); err != nil {
		errCh <- err
		return
	}

	// Step 4: Apply the Crossplane configuration using kubectl apply
	if err := utils.ApplyCrossplaneConfig(instanceTemplateFilename); err != nil {
		errCh <- fmt.Errorf("error applying config for VM %s: %w", vmName, err)
		return // Exit function if applying config fails
	}

	// Step 5: Signal that the processing for this VM is done
	doneCh <- struct{}{}
}

// DeleteCrossplaneConfig deletes Crossplane configuration for the given VM
func DeleteCrossplaneConfig(vmName string) error {
	instanceTemplateFilename := fmt.Sprintf("%s/%s.yaml", instanceTemplateFolder, vmName)
	if err := utils.DeleteCrossplaneConfig(instanceTemplateFilename); err != nil {
		return fmt.Errorf("error deleting Crossplane config: %w", err)
	}
	log.Printf("Crossplane config for VM %s deleted successfully\n", vmName)
	return nil
}
