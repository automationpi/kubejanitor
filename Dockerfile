# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /workspace

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder stage
COPY --from=builder /workspace/manager .

# Copy ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65532:65532

ENTRYPOINT ["/manager"]