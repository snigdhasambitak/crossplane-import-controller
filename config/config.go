package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	knownVMsFile = "known_vms.json"
)

// KnownVMs represents the known VMs
type KnownVMs struct {
	VMs map[string]struct{}
}

// NewKnownVMs creates a new instance of KnownVMs
func NewKnownVMs() *KnownVMs {
	return &KnownVMs{VMs: make(map[string]struct{})}
}

// Save saves the known VMs to a file
func (kv *KnownVMs) Save() error {
	// Create a map to hold the data to be marshalled
	data := make(map[string]interface{})
	for vm := range kv.VMs {
		// Store each VM name as a key with an empty value
		data[vm] = struct{}{}
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling known VMs: %w", err)
	}
	err = ioutil.WriteFile(knownVMsFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing known VMs to file: %w", err)
	}
	return nil
}

// LoadKnownVMs loads the known VMs from a file
func LoadKnownVMs() (*KnownVMs, error) {
	data, err := ioutil.ReadFile(knownVMsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, return a new instance of KnownVMs
			return NewKnownVMs(), nil
		}
		return nil, fmt.Errorf("error reading known VMs file: %w", err)
	}
	if len(data) == 0 {
		// File is empty, return a new instance of KnownVMs
		return NewKnownVMs(), nil
	}

	// Create a map to hold the unmarshalled data
	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling known VMs: %w", err)
	}

	// Extract VM names from the map
	knownVMs := NewKnownVMs()
	for vm := range jsonData {
		knownVMs.Add(vm)
	}

	return knownVMs, nil
}

// Add adds a new VM to the known VMs
func (kv *KnownVMs) Add(vmName string) bool {
	if _, ok := kv.VMs[vmName]; !ok {
		kv.VMs[vmName] = struct{}{}
		// Save known VMs to file
		if err := kv.Save(); err != nil {
			log.Println("Error saving known VMs:", err)
		}
		return true
	}
	return false
}

// Remove removes a VM from the known VMs
func (kv *KnownVMs) Remove(vmName string) {
	delete(kv.VMs, vmName)
	// Save known VMs to file
	if err := kv.Save(); err != nil {
		log.Println("Error saving known VMs:", err)
	}
}

// Get returns the known VMs
func (kv *KnownVMs) Get() map[string]struct{} {
	return kv.VMs
}
