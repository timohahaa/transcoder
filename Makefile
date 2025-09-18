#BUILD_ENVPARMS:=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
BUILD_ENVPARMS:=GOOS=darwin GOARCH=arm64 CGO_ENABLED=0
LDFLAGS:=-s -w

build-composer:
	@[ -d .build ] || mkdir -p .build
	@$(BUILD_ENVPARMS) go build -ldflags "$(LDFLAGS)" -o .build/composer cmd/composer/main.go
	@file  .build/composer
	@du -h .build/composer

build-worker:
	@[ -d .build ] || mkdir -p .build
	@$(BUILD_ENVPARMS) go build -ldflags "$(LDFLAGS)" -o .build/worker cmd/worker/main.go
	@file  .build/worker
	@du -h .build/worker

.PHONY: build
build: build-composer build-worker


.PHONY: proto
proto:
	@echo "Compiling proto files"
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/composer/*.proto

run-composer:
	@DEBUG=true \
	REDIS_ADDRS=localhost:6767 \
	REDIS_USERNAME=redis-user \
	REDIS_PASSWORD=redis-pass \
	POSTGRES_DSN=postgres://postgres:password@0.0.0.0:5432/main?sslmode=disable \
	go run cmd/composer/main.go

run-worker:
	@DEBUG=true \
	COMPOSER_ADDRS=localhost:9999 \
	go run cmd/worker/main.go


compose-up:
	@echo "Stop current dev environment..."
	docker compose down
	@echo "Setup dev environment..."
	docker compose up --build

compose-down:
	docker compose down

compose-shutdown:
	docker compose down --volumes

