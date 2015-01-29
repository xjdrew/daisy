#!/usr/bin/env bash

# dependences: protoc-gen-go 
# install: go get -u github.com/golang/protobuf/protoc-gen-go
# The compiler plugin, protoc-gen-go, will be installed in $GOBIN,
# defaulting to $GOPATH/bin.  It must be in your $PATH for the protocol
# compiler, protoc, to find it.

echo "------------- compile protobuffer file -------------"
set -xe
PB_IN_DIR=$PWD/contrib/proto
PB_OUT_DIR=$PWD/pb

# convert *.proto to *.pb.go
PROTO_FILES=
cd $PB_IN_DIR
subdirs=`find -L . -type d -printf '%P\n'`
for dir in $subdirs;do
    proto_file=$(find $dir -name '*.proto')
    if [ "$proto_file" ];then
        protoc --go_out=$PB_OUT_DIR $proto_file
        PROTO_FILES=${PROTO_FILES}" "${proto_file}
    fi
done
cd -

# convert import XXX "path/to/file.pb" to import XXX "github.com/xjdrew/daisy/pb/path"
PREFIX="github.com/xjdrew/daisy/pb/"
function normalize_import() {
    if [ -z "$1" ];then
        return
    fi

    local file=$1
    local tmpfile=${file}.old
    for i in ${PROTO_FILES};do
        local src=$(echo $i | sed 's/proto$/pb/')
        local dst=${PREFIX}$(echo $src | sed 's/\/.*//')
        mv $file $tmpfile
        sed "/^import/s;$src;$dst;" $tmpfile > $file
        rm $tmpfile
    done
}

GO_FILES=$(find $PB_OUT_DIR -name "*.pb.go")
for src in ${GO_FILES};do
    normalize_import ${src}
done

set +xe
echo "----------------------- end -----------------------"
