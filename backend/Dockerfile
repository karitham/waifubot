FROM golang:1.25-alpine AS builder

ENV GOEXPERIMENT="jsonv2"
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o /go/bin/waifubot ./cmd/waifubot

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/waifubot /bin/

CMD ["/bin/waifubot"]
