**Crossplane Import Controller**

This application is designed to continuously monitor for new virtual machines (VMs) created in Google Cloud Platform (GCP), generate Crossplane configuration files for these VMs, and apply them to a Kubernetes cluster.

**Use Case**

When deploying virtual machines in GCP, it's often necessary to manage their configurations and resources using Kubernetes. Crossplane is a Kubernetes add-on that allows you to define and manage cloud infrastructure using Kubernetes-style APIs. This application automates the process of importing GCP VMs into Kubernetes using Crossplane.

**Prerequisites**

- **GCP Project**: Ensure you have a Google Cloud Platform project set up with VM instances that you want to import into Kubernetes.
- **Kubernetes Cluster**: Set up a Kubernetes cluster where you want to manage the imported VMs using Crossplane.
- **kubectl**: Install `kubectl` CLI tool to interact with the Kubernetes cluster.
- **Docker**: Install Docker to build and run the application inside containers.

**Running the Application Locally**

1. **Clone Repository**: Clone this repository to your local machine.

   ```
   git clone <repository-url>
   ```

2. **Set Environment Variables**: Set the `GCP_PROJECT_ID` environment variable to your GCP project ID.

   ```
   export GCP_PROJECT_ID=<your-project-id>
   ```

3. **Build and Run the Application**: Run the application using the Go compiler.

   ```
   go run main.go
   ```

   The application will continuously monitor for new VMs in the specified GCP project, generate Crossplane configurations, and apply them to the Kubernetes cluster.

**Running the Application with Docker**

1. **Build Docker Image**: Build a Docker image for the application using the provided Dockerfile.

   ```
   docker build -t crossplane-import-controller .
   ```

2. **Run Docker Container**: Run a Docker container with the built image, ensuring to pass the `GCP_PROJECT_ID` environment variable.

   ```
   docker run -e GCP_PROJECT_ID=<your-project-id> -p 8080:8080 crossplane-import-controller
   ```

   Replace `<your-project-id>` with your actual GCP project ID.

**Configuration**

- **Instance Templates**: The application expects instance configuration templates to be stored in the `config` folder. Each template should be named `instance_template.yaml` and contain placeholders for VM-specific configurations.
- **Crossplane Configuration**: The Crossplane configuration files generated by the application will be stored in the `instanceTemplates` folder.
- **Known VM JSON File**: The application maintains a JSON file `known_vms.json` in the root directory to save the state of known VMs. This file is used to track existing VMs and ensure that they are properly managed.

**Logs**

```go
2024/05/09 01:23:59 Fetching VM names from GCP...
2024/05/09 01:24:03 Fetched VM names: [gke-playground-commo-default-e2-stand-e4ac95ff-010p gke-playground-commo-tooling-e2-stand-1001f6aa-5g0c gke-playground-commo-tooling-e2-stand-1001f6aa-vqod gke-playground-commo-tooling-e2-stand-1001f6aa-xinv]
2024/05/09 01:24:03 Applying Crossplane config for VM gke-playground-commo-default-e2-stand-e4ac95ff-010p...
2024/05/09 01:24:07 Applying Crossplane config for VM gke-playground-commo-tooling-e2-stand-1001f6aa-5g0c...
2024/05/09 01:24:08 Applying Crossplane config for VM gke-playground-commo-tooling-e2-stand-1001f6aa-vqod...
2024/05/09 01:24:09 Applying Crossplane config for VM gke-playground-commo-tooling-e2-stand-1001f6aa-xinv...
2024/05/09 01:26:11 Fetching VM names from GCP...
2024/05/09 01:26:13 Fetched VM names: [gke-playground-commo-default-e2-stand-e4ac95ff-010p gke-playground-commo-tooling-e2-stand-1001f6aa-5g0c gke-playground-commo-tooling-e2-stand-1001f6aa-vqod gke-playground-commo-tooling-e2-stand-1001f6aa-xinv instance-20240508-232441]
2024/05/09 01:26:13 Applying Crossplane config for VM instance-20240508-232441...
2024/05/09 01:28:16 Fetching VM names from GCP...
2024/05/09 01:28:19 Fetched VM names: [gke-playground-commo-default-e2-stand-e4ac95ff-010p gke-playground-commo-tooling-e2-stand-1001f6aa-5g0c gke-playground-commo-tooling-e2-stand-1001f6aa-vqod gke-playground-commo-tooling-e2-stand-1001f6aa-xinv instance-20240508-232441]
2024/05/09 01:30:19 Fetching VM names from GCP...
2024/05/09 01:30:23 Fetched VM names: [gke-playground-commo-default-e2-stand-e4ac95ff-010p gke-playground-commo-tooling-e2-stand-1001f6aa-5g0c gke-playground-commo-tooling-e2-stand-1001f6aa-vqod gke-playground-commo-tooling-e2-stand-1001f6aa-xinv]
2024/05/09 01:30:23 Deleting Crossplane config for VM instance-20240508-232441...
2024/05/09 01:30:25 Crossplane config for VM instance-20240508-232441 deleted successfully
^Csignal: interrupt
```


**Contributing**

Contributions are welcome! If you encounter any issues or have suggestions for improvements, please feel free to open an issue or submit a pull request.

**License**

This project is licensed under the [MIT License](LICENSE).