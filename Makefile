BUILD_ENVPARMS:=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
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
    proto/transcoder/*.proto


run-composer:
	@DEBUG=true \
	go run cmd/composer/main.go

run-worker:
	@DEBUG=true \
	go run cmd/worker/main.go
