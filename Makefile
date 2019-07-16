all: build

build: test
	go build

test:
	go test -race -v -vet all

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out