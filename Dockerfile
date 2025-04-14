# Stage 1: Build the Go application
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Caching is leveraged to speed up builds.
RUN go mod download

# Copy the source files from the current directory to the working directory
ADD ["internal/", "./internal/"]
ADD ["pkg/", "./pkg/"]
COPY ["cmd/main.go", "./"]

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on (if applicable)
EXPOSE 8081

# Command to run the executable
CMD ["./main"]