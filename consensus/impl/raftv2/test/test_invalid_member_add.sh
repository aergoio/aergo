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
echo "========= add invalid config member aergo4 ========="
make_node.sh
sleep 3

pushd $TEST_RAFT_INSTANCE/config
do_sed.sh BP11004.toml 13002 13009 =
popd

add_member.sh aergo4

sleep 20
existProcess 10004
if [ "$?" = "1" ]; then
	echo "error! process must be killed."
	exit
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
