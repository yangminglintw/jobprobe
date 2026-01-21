# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o mock-api .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/mock-api /usr/local/bin/mock-api

ENV PORT=8000

EXPOSE 8000

CMD ["mock-api"]
