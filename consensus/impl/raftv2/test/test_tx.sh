#!/usr/bin/env bash
#./aergoconf-gen.sh 10001 tmpl.toml 5
echo "============================== raft tx test ============================"
source set_test_env.sh
source test_common.sh

echo ""
echo "======== make initial server ========="
make_node.sh 

checkSync 10001 10002 60
checkSync 10001 10003 60

pushd  $TEST_RAFT_INSTANCE_CLIENT
run_tx.sh
popd

checkSync 10001 10002 60
checkSync 10001 10003 60


