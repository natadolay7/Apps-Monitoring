# =========================
# Stage 1: Build
# =========================
FROM golang:1.25.2-alpine AS builder

WORKDIR /app

# Copy dependency dulu (cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

ENV PORT=8282
ENV GIN_MODE=release

EXPOSE 8282

CMD ["./app"]
