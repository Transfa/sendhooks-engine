# Use the official Go image from the DockerHub
FROM golang:1.20 as builder

# We set the Current Working Directory inside the container
WORKDIR /app

# We copy everything from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook .

#### Start a new stage from scratch ####
FROM alpine:latest  

# Install CA certificates as we will need it when adding the Redis SSL configuration
RUN apk --no-cache add ca-certificates 

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/webhook .

# Command to run the executable
CMD ["./webhook"]
