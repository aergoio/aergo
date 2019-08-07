#!/usr/bin/env bash
echo "============================== raft server boot & down test ============================"

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh
source test_common.sh

echo ""
echo "======== make initial server ========="
make_node.sh 

checkSync 10001 10002 60
checkSync 10001 10003 60


for ((idx=0; idx<=2; idx++)); do
	echo "try $idx"
	echo "======== shutdown aergo1 ============"
	kill_svr.sh 11001
	sleep 3
	checkSync 10002 10003 60

	echo "======== restart aergo1 ============"
	run_svr.sh 11001
	sleep 2
	checkSync 10001 10003 60


	echo "======== shutdown aergo2 ============"
	kill_svr.sh 11002
	sleep 3
	checkSync 10001 10003 60

	echo "======== restart aergo2 ============"
	run_svr.sh 11002
	checkSync 10001 10002 60
done
