# Use the official Golang image as the base image
FROM golang:1.20 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o wa-serv main.go

# Use a minimal base image for the final container
FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/wa-serv .

# Expose the port the service will run on
EXPOSE 8080

# Command to run the application
CMD ["./wa-serv"]
