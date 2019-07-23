#!/usr/bin/env bash
echo "================= sync crash of joinned node and restart (nobackup) test ===================="

BP_NAME=""
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

#4 node add & crash
DEBUG_SYNCER_CRASH=1 add_member.sh aergo4

WaitShutdown 11004 30
if [ "$?" = "0" ]; then
	echo "failed to syncer crash of joined node 11004"
	exit 100
fi

# 4 node restart
run_svr.sh 11004
checkSync 10001 10004 120 result
