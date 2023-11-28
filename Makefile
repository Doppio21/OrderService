BINARY_NAME=bin/

all: build
 
build:
	go build -o $(BINARY_NAME) ./cmd/...

lint:
	golangci-lint run
 
test:
	CGO_ENABLED=1 go test ./... -race

gen:
	go generate ./...
