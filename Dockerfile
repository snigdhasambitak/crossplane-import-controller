# Use the official Go image as the base image
FROM golang:1.21-alpine as builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o crossplane-import-controller .

# Use a minimal base image to reduce the image size
FROM google/cloud-sdk:428.0.0-alpine

# Set environment variables
ENV PORT=8080 \
    GCP_PROJECT_ID="playground-common-cros1"

# Install gcloud and kubectl
RUN apk add --update --upgrade  --no-cache \
    python3 \
    py3-pip \
    git \
    && pip3 install \
    google-auth \
    google-api-python-client \
    && gcloud components install gke-gcloud-auth-plugin \
    kubectl \
    && rm -rf google-cloud-sdk/bin/anthoscli \
    && rm -rf /var/cache/apk/*

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/crossplane-import-controller /app/crossplane-import-controller

# Expose the port on which the application will run
EXPOSE $PORT

# Set the entry point for the container
ENTRYPOINT ["/app/crossplane-import-controller"]
