# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for version info
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/user/jobprobe/cmd.Version=${VERSION} \
              -X github.com/user/jobprobe/cmd.Commit=${COMMIT} \
              -X github.com/user/jobprobe/cmd.BuildDate=${BUILD_DATE}" \
    -o jprobe .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/jprobe /usr/local/bin/jprobe

# Create config directory
RUN mkdir -p /etc/jprobe

# Default config path
ENV JPROBE_CONFIG=/etc/jprobe

ENTRYPOINT ["jprobe"]
CMD ["run", "--config", "/etc/jprobe"]
