.PHONY: all test

all:
	go build

test:
	go test -v ./...
