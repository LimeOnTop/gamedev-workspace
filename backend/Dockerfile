# syntax=docker/dockerfile:1.4

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o workspace-service ./cmd

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata wget \
	&& adduser -D -s /bin/sh appuser \
	&& mkdir -p /data/uploads \
	&& chown -R appuser:appuser /data

WORKDIR /app

COPY --from=builder /app/workspace-service .

RUN chown -R appuser:appuser /app

USER appuser

ENV UPLOAD_DIR=/data/uploads PORT=8080

EXPOSE 8080
CMD ["./workspace-service"]
