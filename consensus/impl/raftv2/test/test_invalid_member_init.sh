#!/usr/bin/env bash
echo "================= raft invalid member test ===================="

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh
source test_common.sh

rm BP11004* BP11005*

echo "kill_svr & clean 11004~11007"
kill_svr.sh
for i in  11004 11005 11006 11007; do
	rm -rf ./data/$i
	rm -rf ./BP$i.toml
done


echo ""
echo "========= invalid initial member node3  ========="
pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11001.toml aergo2 aergo2_xxx =
popd

make_node.sh
sleep 3

HasLeader 10001 result
if [ "$result" = "1" ]; then
	echo "failed to verify invalid cluster"
	exit
fi

checkSync 10002 10003 30 result

pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11001.toml aergo2_xxx aergo2 =
popd

#
#echo ""
#echo "========= check if reorg occured ======="
#checkReorg
