#!/usr/bin/env bash
source test_common.sh

echo ""
echo "============================== raft server slow follower node test ============================"

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh

echo ""
echo "======== make initial server ========="
make_node.sh 

checkSync 10001 10002 30
checkSync 10001 10003 30

sleep 10

kill_svr.sh 11003

echo "run aergo3(11003). this node is slower than other nodeds."
DEBUG_CHAIN_OTHER_SLEEP=10000 run_svr.sh 11003

checkSyncRunning 10001 10003 20

echo "------------ success--------------"


