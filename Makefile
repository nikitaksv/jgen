
GOPATH:=$(shell go env GOPATH)
.PHONY: init
init:
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get github.com/micro/micro/v3/cmd/protoc-gen-micro
.PHONY: proto
proto:
	protoc -I=$(GOPATH)/src/github.com/googleapis --proto_path=. --micro_out=. --go_out=:. proto/gendata.proto
	
.PHONY: build
build:
	go build -o gendata *.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t gendata:latest
