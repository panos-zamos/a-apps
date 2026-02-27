FROM golang:1.21-alpine AS builder

RUN apk --no-cache add build-base

WORKDIR /app

# Copy shared modules and app manifests (context: repo root)
COPY shared ./shared
COPY apps/projects/go.mod apps/projects/go.sum ./apps/projects/

WORKDIR /app/apps/projects
RUN go mod download

# Copy source code
COPY apps/projects ./

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/apps/projects/main .
COPY --from=builder /app/apps/projects/config.yaml .

# Create data directory
RUN mkdir -p data

EXPOSE 3002

CMD ["./main"]
