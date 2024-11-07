# Use the official Go image with the required version
FROM golang:1.23-alpine

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/webapp

# Expose the port the app runs on
EXPOSE 8080

# Run the executable
CMD ["./main"]
