# Build stage for frontend assets
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package.json files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source files
COPY . .

# Build Tailwind CSS
RUN npm run build-css

# Build stage for Go application
FROM golang:1.24-alpine AS go-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non root user
RUN addgroup -S -g 1600 transfer && adduser -S -u 1600 transfer -G transfer

WORKDIR /app

# Copy the Go binary from builder stage
COPY --from=go-builder /app/main .

# Copy static files and templates
COPY --from=frontend-builder /app/static ./static
COPY --from=frontend-builder /app/templates ./templates

# Expose port
EXPOSE 42069

# Run the application
CMD ["./main"]
