FROM golang:1.26-alpine AS builder

ENV GOEXPERIMENT="jsonv2"
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" -o /go/bin/waifubot ./cmd/waifubot

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/waifubot /bin/

CMD ["/bin/waifubot"]
