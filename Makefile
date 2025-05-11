all: build test

build:
	@go build -o bingwp main.go
test:
	@go test -v ./...
cover:
	@go test -coverprofile cover.out ./... && go tool cover -func=cover.out && rm cover.out
