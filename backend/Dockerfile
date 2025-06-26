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

COPY --from=builder /go/bin/bot /bin/
COPY --from=builder /go/bin/api /bin/

CMD ["/bin/bot"]
