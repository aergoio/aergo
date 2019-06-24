#!/usr/bin/env bash

source test_common.sh

echo ""
echo "============================== raft server slow leader node test ============================"
echo "description: run aergo. Chain service of leader node works slowly."
echo ""

make_node.sh

kill_svr.sh

DEBUG_CHAIN_BP_SLEEP=10000 run_svr.sh
sleep 5

# raft에 의해 leader가 바뀌지 않아야 한다. 
isStableLeader 20
ret=$?
if [ "$ret" == "0" ];then
	echo "=========== failed =========================="
	exit 100
fi
echo "=========== succeed =========================="
