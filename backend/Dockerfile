# Use the official Golang image with Go 1.20 as a build environment
FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files from the backend directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire backend source code into the container
COPY . ./

# Set the working directory to where main.go is located
WORKDIR /app/cmd/webapp

# Build the Go app
RUN go build -o /app/main .

# Use a minimal base image
FROM gcr.io/distroless/base-debian11

# Set the working directory inside the container
WORKDIR /

# Copy the compiled Go binary from the builder image
COPY --from=builder /app/main .

# Copy the service account key if necessary
COPY --from=builder /app/serviceAccountKey.json ./backend/serviceAccountKey.json

# Expose port 8080
EXPOSE 8080

# Run the Go app
ENTRYPOINT ["/main"]
