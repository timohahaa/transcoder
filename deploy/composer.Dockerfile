########## BUILD STAGE ##########
FROM golang:1.25.1 AS builder

WORKDIR /src

# deps
COPY go.mod .
RUN go mod download

# source code
COPY . .

# compile
RUN make build-composer

########## RUN STAGE ##########
FROM alpine:latest

RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY --from=builder /src/.build/composer ./composer

ENTRYPOINT [ "./composer" ]
