# ---------- Build stage ----------
FROM golang:1.25 AS builder

WORKDIR /app

# download dependencies for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the code
COPY . .

# Build static binary for Linux (K8s container runtime)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pod-killer ./cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/pod-killer /app/pod-killer

# Container runs as non-root user
USER nonroot:nonroot

ENTRYPOINT ["/app/pod-killer"]