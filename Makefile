.PHONY: all test lint

all: todo/todo.pb.go
	go build

test: todo/todo.pb.go
	go test -v ./...

lint: todo/todo.pb.go
	golint ./...
	errcheck ./...

todo/todo.pb.go: proto/todo.proto
	cd proto; protoc -I. \
		-I${GATEWAY}/third_party/googleapis \
		--go_out=plugins=grpc,paths=source_relative:../todo \
		--grpc-gateway_out=logtostderr=true,paths=source_relative:../todo \
		todo.proto
