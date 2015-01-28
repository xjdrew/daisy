.PHONY: all install test clean

all: protobuf install

%.pb.go: %.proto
	protoc --go_out=$(PB_DIR) $<

protobuf:
	@project/pb-gen.sh

install:
	go install .

test:
	go test ./protolist

clean:
	go clean -i ./...
