FROM golang:1.25-alpine AS builder

ENV GOEXPERIMENT="jsonv2"
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o /go/bin/bot ./cmd/bot
RUN go build -ldflags="-w -s" -o /go/bin/api ./cmd/api

FROM alpine:latest

ARG TARGETARCH
RUN apk add --no-cache curl postgresql-client && \
    curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-${TARGETARCH} && \
    chmod +x /usr/local/bin/dbmate

COPY --from=builder /go/bin/bot /bin/
COPY --from=builder /go/bin/api /bin/
COPY db/migrations /migrations

COPY <<'EOF' /entrypoint.sh
#!/bin/sh
set -e
if [ -n "$$DB_URL" ]; then
  echo "Running database migrations..."
  dbmate --env DB_URL --migrations-dir /migrations --no-dump-schema up
fi
exec "$@"
EOF

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/bin/bot"]
