FROM golang:1.24-alpine AS builder

RUN apk add git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /build/bot ./cmd/bot
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /build/api ./cmd/api

FROM alpine:3.20
COPY --from=builder /build /bin

CMD ["/bin/bot"]
