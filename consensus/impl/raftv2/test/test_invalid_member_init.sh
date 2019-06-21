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
do_sed.sh BP11001.toml aergo2 aergo2_xxx =
popd

TEST_SKIP_GENESIS=0 make_node.sh 
RUN_TEST_SCRIPT set_system_admin.sh

HasLeader 10001 result
if [ "$result" = "1" ]; then
	echo "failed to verify invalid cluster"
	exit 100
fi

checkSync 10002 10003 30 result

pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11001.toml aergo2_xxx aergo2 =
popd

#
#echo ""
#echo "========= check if reorg occured ======="
#checkReorg
