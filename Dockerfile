# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /loxone-bridge ./cmd/loxone-bridge

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /loxone-bridge /usr/local/bin/loxone-bridge

EXPOSE 8080

ENTRYPOINT ["loxone-bridge"]
