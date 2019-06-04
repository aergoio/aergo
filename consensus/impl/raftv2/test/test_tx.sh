#!/usr/bin/env bash
#./aergoconf-gen.sh 10001 tmpl.toml 5
echo "============================== raft tx test ============================"
source set_test_env.sh
source test_common.sh

echo "pushd TEST_RAFT_INSTANCE=$TEST_RAFT_INSTANCE"

pushd $TEST_RAFT_INSTANCE
clean.sh all
rm BP*

prepareConfig 3

echo ""
echo "======== make initial server ========="
make_node.sh 

checkSync 10001 10002 30
checkSync 10001 10003 30
popd

pushd  $TEST_RAFT_INSTANCE_CLIENT
run_tx.sh
popd
checkSync 10001 10002 30
checkSync 10001 10003 30


