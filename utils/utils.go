package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

// ReadConfigFile reads a configuration file
func ReadConfigFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", path, err)
	}
	return string(content), nil
}

// WriteConfigFile writes content to a file
func WriteConfigFile(content, filename string) error {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing config to file %s: %w", filename, err)
	}
	return nil
}

// ReplacePlaceholder replaces a placeholder in a string with a value
func ReplacePlaceholder(str, placeholder, value string) string {
	return strings.ReplaceAll(str, placeholder, value)
}

// ApplyCrossplaneConfig applies Crossplane configuration using kubectl apply
func ApplyCrossplaneConfig(filename string) error {
	cmd := exec.Command("kubectl", "apply", "-f", filename)
	log.Printf("Applying Crossplane config from file %s...\n", filename)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error applying Crossplane config: %w", err)
	}
	return nil
}

// DeleteCrossplaneConfig deletes Crossplane configuration using kubectl delete
func DeleteCrossplaneConfig(filename string) error {
	cmd := exec.Command("kubectl", "delete", "-f", filename)
	log.Printf("Deleting Crossplane config from file %s...\n", filename)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error deleting Crossplane config: %w", err)
	}
	return nil
}
