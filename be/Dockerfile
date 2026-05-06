FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrator ./cmd/migrate

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server
COPY --from=builder /migrator /migrator
COPY migrations /migrations

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
