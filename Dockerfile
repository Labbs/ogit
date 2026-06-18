FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/app ./cmd

FROM alpine:latest AS release
RUN apk update && apk add ca-certificates git && rm -rf /var/cache/apk/*
COPY --from=builder /app/bin/app .
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]