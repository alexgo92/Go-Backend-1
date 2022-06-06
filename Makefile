.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	@echo "Start test"
	go test -covermode=atomic ./...

.PHONY: build
build:
	@echo "Start build"
	GOOS=linux
	go build -o bin/myapp

.PHONY: run
run:
	bin/myapp

.PHONY: all
all: lint test build
	@echo "Success"
