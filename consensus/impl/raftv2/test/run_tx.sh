#!/usr/bin/env bash
source set_test_env.sh

pushd $TEST_RAFT_INSTANCE_CLIENT

port=$1
if [ "$port" == "" ]; then
    port=10001
fi

txPerAcc=$2
if [ "$txPerAcc" == "" ]; then
    txPerAcc=10
fi

generate_tx.sh $port $txPerAcc

commit_tx.sh

popd
