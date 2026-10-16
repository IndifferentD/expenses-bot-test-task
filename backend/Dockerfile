FROM golang:1.26.8-alpine3.24 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p /out && CGO_ENABLED=0 go build -trimpath -o /out/ ./cmd/...

FROM alpine:3.24

RUN apk add --no-cache ca-certificates
COPY --from=builder /out/ /usr/local/bin/
