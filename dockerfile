# Use the official Golang image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files first (for dependency caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN go build -o main ./cmd/.

# Expose the application port (replace 3333 with your app's port if different)
EXPOSE 3333

# Run the application
CMD ["./main"]
