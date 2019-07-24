#!/usr/bin/env bash
echo "================= raft new cluster with backup test ===================="

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh
source test_common.sh

rm -rf  $TEST_RAFT_INSTANCE_DATA
rm BP*.toml

echo "start new cluster for bakcup"
TEST_SKIP_GENESIS=0 make_node.sh
sleep 2

# kill aergo3
kill_svr.sh 

# prepare backup data
for i in 11001 11002; do 
	rm -rf  $TEST_RAFT_INSTANCE/data/$i
	cp -rf $TEST_RAFT_INSTANCE/data/11003 $TEST_RAFT_INSTANCE/data/$i
done

do_sed.sh "toml" "usebackup=false" "usebackup=true" ":"
run_svr.sh 11001

RUN_TEST_SCRIPT set_system_admin.sh

date
echo ""
echo "========= add aergo4 ========="
add_member.sh aergo4
checkSync 10001 10004 120 result
