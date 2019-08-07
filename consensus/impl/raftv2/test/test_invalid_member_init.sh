#!/usr/bin/env bash
echo "================= raft invalid member init test ===================="

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh
source test_common.sh


echo "kill_svr"
kill_svr.sh
rm -rf $TEST_RAFT_INSTANCE_DATA
rm $TEST_RAFT_INSTANCE/BP*

echo ""
echo "========= invalid initial member node3  ========="
pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11001.toml aergo1 aergo1_xxx =

TEST_SKIP_GENESIS=0 TEST_NOWAIT_PEER=1 make_node.sh
RUN_TEST_SCRIPT set_system_admin.sh

sleep 10

existProcess 10001
if [ "$?" = "1" ]; then
	echo "failed to verify invalid cluster"
	exit 100
fi

echo "node aergo1 is crashed because of invalid config"

checkSync 10002 10003 60 result

pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11001.toml aergo1_xxx aergo1 =
popd


echo ""
echo "========= check if reorg occured ======="
checkReorg
