FROM golang:1.21-alpine AS builder

RUN apk --no-cache add build-base

WORKDIR /app

# Copy shared modules and app manifests (context: repo root)
COPY shared ./shared
COPY apps/todo-list/go.mod apps/todo-list/go.sum ./apps/todo-list/

WORKDIR /app/apps/todo-list
RUN go mod download

# Copy source code
COPY apps/todo-list ./

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/apps/todo-list/main .
COPY --from=builder /app/apps/todo-list/config.yaml .

# Create data directory
RUN mkdir -p data

EXPOSE 3001

CMD ["./main"]
