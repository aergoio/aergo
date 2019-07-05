#!/usr/bin/env bash
echo "================= raft invalid member add test ===================="

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
echo "========= join invalid config member aergo4 ========="
pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11004.toml 127 137 =
popd

TEST_SKIP_GENESIS=0 make_node.sh 
RUN_TEST_SCRIPT set_system_admin.sh

add_member.sh aergo4
if [ $? -ne 0 ];then
	echo "Adding of invalid config member must succeed"
	exit 100
fi

sleep 40
existProcess 10004
if [ "$?" = "1" ]; then
	echo "error! process must be killed."
	exit 100
fi

pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11004.toml 13009 13002 =
popd

echo ""
echo "========= rm aergo4 ========="
rm_member.sh aergo4
rm BP11004*
checkSync 10001 10003 20 

echo ""
echo "========= check if reorg occured ======="
checkReorg
