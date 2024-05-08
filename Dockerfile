# Use the official Golang image to build the application
FROM golang:1.21.6 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Use a minimal Alpine image as the base image for the final container
FROM alpine:latest

# Set environment variables
ENV GCP_PROJECT_ID="playground-common-cros1"
ENV PORT="8080"

# Copy the binary built in the previous stage into the container
COPY --from=build /app/app /app/app

# Expose the port on which the application will listen
EXPOSE 8080

# Command to run the application
CMD ["/app/app"]
