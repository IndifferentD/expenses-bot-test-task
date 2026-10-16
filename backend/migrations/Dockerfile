FROM golang:1.26.8-alpine3.24 AS builder

RUN CGO_ENABLED=0 go install \
    -tags='no_clickhouse no_mssql no_mysql no_sqlite3 no_ydb no_vertica no_libsql no_azuresql' \
    github.com/pressly/goose/v3/cmd/goose@v3.29.0

FROM alpine:3.24

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
ENTRYPOINT ["goose", "-dir", "/migrations"]
