#BUILD_ENVPARMS:=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
BUILD_ENVPARMS:=GOOS=darwin GOARCH=arm64 CGO_ENABLED=0
LDFLAGS:=-s -w

build-composer:
	@[ -d .build ] || mkdir -p .build
	@$(BUILD_ENVPARMS) go build -ldflags "$(LDFLAGS)" -o .build/composer cmd/composer/main.go
	@file  .build/composer
	@du -h .build/composer

build-encoder:
	@[ -d .build ] || mkdir -p .build
	@$(BUILD_ENVPARMS) go build -ldflags "$(LDFLAGS)" -o .build/encoder cmd/encoder/main.go
	@file  .build/encoder
	@du -h .build/encoder

.PHONY: build
build: build-composer build-encoder


.PHONY: proto
proto:
	@echo "Compiling proto files"
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/composer/*.proto

run-composer:
	@DEBUG=true \
	REDIS_ADDRS=localhost:6379 \
	REDIS_USERNAME=default \
	REDIS_PASSWORD=redis-pass \
	POSTGRES_DSN=postgres://postgres:password@0.0.0.0:5432/main?sslmode=disable \
	HTTP_ADDR=localhost:8080 \
	GRPC_ADDR=localhost:9090 \
	go run cmd/composer/main.go

run-encoder:
	@DEBUG=true \
	COMPOSER_ADDRS=localhost:9090 \
	go run cmd/encoder/main.go


compose-up:
	@echo "Stop current dev environment..."
	docker compose down
	@echo "Setup dev environment..."
	docker compose up --build

compose-down:
	docker compose down

compose-shutdown:
	docker compose down --volumes

