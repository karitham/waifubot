FROM golang:1.27-alpine AS builder

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s -X main.version=${VERSION}" -o /go/bin/waifubot ./cmd/waifubot

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/waifubot /bin/

CMD ["/bin/waifubot"]
