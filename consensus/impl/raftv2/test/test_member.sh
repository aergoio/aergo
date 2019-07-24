#!/usr/bin/env bash
echo "================= raft member add/remove test ===================="

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

TEST_SKIP_GENESIS=0 make_node.sh
RUN_TEST_SCRIPT set_system_admin.sh

date
echo ""
echo "========= add aergo4 ========="
add_member.sh aergo4
if [ "$?" != "0" ];then
	echo "fail to add"
	exit 100
fi
checkSync 10001 10004 120 result

date
echo ""
echo "========= add aergo5 ========="
add_member.sh aergo5
if [ "$?" != "0" ];then
	echo "fail to add"
	exit 100
fi
checkSync 10001 10005 120 


date
echo ""
echo "========= add aergo6 ========="
add_member.sh aergo6
if [ "$?" != "0" ];then
	echo "fail to add"
	exit 100
fi
checkSync 10001 10006 120 result


date
echo ""
echo "========= add aergo7 ========="
add_member.sh aergo7
if [ "$?" != "0" ];then
	echo "fail to add"
	exit 100
fi
checkSync 10001 10007 120 

date
echo ""
echo "========= rm aergo7 ========="
rm_member.sh aergo7
if [ "$?" != "0" ];then
	echo "fail to rm"
	exit 100
fi
rm BP11007*
checkSync 10001 10006 60 

echo ""
echo "========= rm aergo6 ========="
rm_member.sh aergo6
if [ "$?" != "0" ];then
	echo "fail to rm"
	exit 100
fi
rm BP11006*
checkSync 10001 10005 60 

echo ""
echo "========= rm aergo5 ========="
rm_member.sh aergo5
if [ "$?" != "0" ];then
	echo "fail to rm"
	exit 100
fi
rm BP11005*
checkSync 10001 10004 60 

echo ""
echo "========= rm aergo4 ========="
rm_member.sh aergo4
if [ "$?" != "0" ];then
	echo "fail to rm"
	exit 100
fi
rm BP11004*
checkSync 10001 10003 60 

echo ""
echo "========= check if reorg occured ======="
checkReorg
