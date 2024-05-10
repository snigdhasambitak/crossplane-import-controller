package cloudprovider

// CloudProvider interface defines methods for interacting with cloud providers
type CloudProvider interface {
	GetVMNames(projectID string) ([]string, error)
}
