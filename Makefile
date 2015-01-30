.PHONY: all install test clean

all: build protolist protobuf install

build:
	-mkdir -p ./gen/proto
	-mkdir -p ./gen/interfaces

%.pb.go: %.proto
	protoc --go_out=$(PB_DIR) $<

protolist:
	cat ./contrib/proto/service.protolist | go run ./pb/generator.go > ./gen/interfaces/protolist.go

protobuf:
	project/pb-gen.sh

install:
	go install ./daisy

test:
	go test ./protolist

clean:
	go clean -i ./...
