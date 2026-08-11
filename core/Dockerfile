FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install dependencies required for some CGO builds if needed, though we use pure Go here mostly
# RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o core-service main.go

FROM alpine:latest

# Create a non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/core-service .

# Use non-root user
USER appuser

EXPOSE 8080

CMD ["./core-service"]
