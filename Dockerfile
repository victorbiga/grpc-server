# Use golang base image for the build stage
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Use a minimal base image for the final stage
FROM gcr.io/distroless/base-debian10

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /

# Expose the port that the server listens on
EXPOSE 50051

# Command to run the server executable
CMD ["/server"]
