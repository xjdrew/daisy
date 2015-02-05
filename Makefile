.PHONY: all install test clean

all: build protobuf descriptor install

build:
	-mkdir -p ./gen/proto
	-mkdir -p ./gen/descriptor

%.pb.go: %.proto
	protoc --go_out=$(PB_DIR) $<

descriptor:
	cat ./contrib/proto/service.protolist | go run ./pb/generator.go > ./gen/descriptor/descriptor.go

protobuf:
	project/pb-gen.sh

install:
	go install ./daisy
	go install ./client

test:
	go test ./...

clean:
	go clean -i ./...
