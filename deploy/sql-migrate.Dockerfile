FROM golang:1.24 AS builder

WORKDIR /src

RUN GOBIN=/src go install github.com/rubenv/sql-migrate/...@latest

COPY ./migrations/dbconfig.yaml dbconfig.yaml
COPY ./migrations/main/ migrations/main

ENTRYPOINT [ "/src/sql-migrate", "up", "-config=/src/dbconfig.yaml", "-env=main" ]
