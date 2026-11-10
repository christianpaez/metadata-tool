# ---------------------------------------------------
# STAGE 1: Builder
# ---------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and certs (needed if fetching private/remote go packages)
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build a completely static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-w -s" \
  -o /app/server .

# ---------------------------------------------------
# STAGE 2: Runtime (Alpine)
# ---------------------------------------------------
FROM alpine:3.21

# Install CA certificates (for SSL/HTTPS), and tzdata
RUN apk add --no-cache \
  ca-certificates \
  tzdata

# Create non-root user for security (Kamal best practice)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy static binary from builder stage
COPY --from=builder /app/server /home/appuser/server

# Expose port (8080 default)
EXPOSE 8080

CMD ["/home/appuser/server"]
