# ==========================================
# STAGE 1: The Builder
# ==========================================
# We use the official Go image based on Alpine Linux to keep the build environment relatively small.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Download dependencies first. 
# We copy go.mod and go.sum separately to leverage Docker's layer caching.
# If dependencies haven't changed, Docker skips downloading them again.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary.
# CGO_ENABLED=0: This is a crucial Go interview concept. It disables Cgo, ensuring 
# the compiled binary is 100% statically linked and has no dependencies on the host OS's C libraries.
# GOOS=linux: Ensures we are compiling for Linux, which is what AWS ECS runs.
RUN CGO_ENABLED=0 GOOS=linux go build -o shortener-api ./main.go


# ==========================================
# STAGE 2: The Runner (Production Image)
# ==========================================
# Alpine is a hyper-lightweight Linux distribution (~5MB).
FROM alpine:latest

# Security Best Practice: Do not run containers as the root user.
# We create a generic "appuser" to run our binary.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage above
COPY --from=builder /app/shortener-api .

# Change ownership of the binary to our non-root user
RUN chown appuser:appgroup /app/shortener-api

# Switch to the non-root user
USER appuser

# Document that this container listens on port 8080
EXPOSE 8080

# The command to execute when the container starts
CMD ["./shortener-api"]