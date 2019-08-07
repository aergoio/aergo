#!/usr/bin/env bash
echo "================= raft member join with backup : best entry test ===================="

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh
source set_test_env.sh
source test_common.sh

# Test scenario
# aergo3 : log 1 ~ 100 : X X X X (305:300) X X X X : 500(snap) 501  XXXX 0 0 0 0 XXXX  0000 0  0 0 0 XXX 0 0 0 0 
#		   aergo4 backup(block 300: term/logno: 307)

echo "clean all prev servers"
echo "kill_svr & clean 11004~11007"
kill_svr.sh
for i in  11004 11005 11006 11007; do
	echo "rm -rf $TEST_RAFT_INSTANCE/data/$i $TEST_RAFT_INSTANCE/BP$i.toml"
	rm -rf $TEST_RAFT_INSTANCE/data/$i $TEST_RAFT_INSTANCE/BP$i.toml
done

TEST_SKIP_GENESIS=0 make_node.sh
# make snap in aergo1
sleep 20
# aergo3 down
kill_svr.sh 11003

echo "========= after kill 10003 ========="
sleep 100

# copy aergo1 to backup for aergo4 
echo ""
echo "========= copy backup : cp -rf $TEST_RAFT_INSTANCE/data/11001 $TEST_RAFT_INSTANCE/data/11004 ========="
cp -rf $TEST_RAFT_INSTANCE/data/11001 $TEST_RAFT_INSTANCE/data/11004

echo "========= after backup for aergo4 ========="
sleep 100

# kill all
kill_svr.sh 

#aergo3 snapshot sync with aergo1
run_svr.sh 
checkSync 10001 10003 180 

# remove aergo1, aergo2
set_system_admin.sh
echo "=========== rm member1 =========="
rm_member.sh aergo1
rm BP11001*

echo "=========== rm member2 =========="
rm_member.sh aergo2
rm BP11002*

# add aergo4 with backup
echo ""
echo "========= add aergo4 ========="
add_member.sh aergo4 usebackup 10003
checkSync 10003 10004 180
#checkSyncWithLeader 10004 180
# check log if "can't find raft entry for requested hash. so try to find closest raft entry." exists in aergo3

egrep -q 'find closest raft entry' $TEST_RAFT_INSTANCE/server_BP11003.log
if [ "$?" != "0" ]; then
	echo "not occure log: find closest raft entry"
	exit 100
fi

echo "succeed to sync"
