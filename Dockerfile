# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs before building
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/api/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /users-api ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /users-api .

EXPOSE 8080

ENTRYPOINT ["/app/users-api"]
