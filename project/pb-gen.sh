#!/usr/bin/env bash
echo "------------- compile protobuffer file -------------"
set -xe
PB_IN_DIR=$PWD/contrib/proto
PB_OUT_DIR=$PWD/pb

subdirs=`cd $PB_IN_DIR && find -L . -type d | sed 's/^\.\///' |grep -v '^\.$'`
for dir in $subdirs;do
    in_dir=$PB_IN_DIR/$dir
    out_dir=$PB_OUT_DIR/$dir
    mkdir -p $out_dir

    proto_file=`cd $in_dir && find . -name '*.proto'`
    if [ $proto_file ];then
        cd $in_dir && protoc --go_out=$out_dir $proto_file && cd - >/dev/null
    fi
done
set +xe
echo "----------------------- end -----------------------"
