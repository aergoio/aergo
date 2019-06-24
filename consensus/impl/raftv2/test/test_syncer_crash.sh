#!/usr/bin/env bash
echo "============================== raft syncer crash test (crash=$method)============================"
source test_common.sh


if [ $# != 1 ];then
	echo "Usage: $0 crashno(0=fatal, 1=error)"
fi

CRASH_NO=$1
method=""
if [ "$CRASH_NO" = "1" ]; then
	method="FATAL"
else 
	method="ERROR"
fi


BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#clean.sh
#./inittest.sh

echo ""
echo "======== make initial server ========="
make_node.sh 

checkSync 10001 10002 10
checkSync 10001 10003 10

echo "kill for delaying 11003"
kill_svr.sh 11003

sleep 20
DEBUG_SYNCER_CRASH=$CRASH_NO run_svr.sh 11003
# aergo3 (11003)은 crash(CRASH_NO=1) or syncer 에러후(CRASH_NO=0) 정상 상태
sleep 10

# leader에서 aergo3의 raftstate가 Snapshot이 아니어야 한다. 
# get leaderport
getLeaderPort leaderport
echo "leaderport=$leaderport"

# get raftid for aergo3
name="aergo3"
raftState=
getRaftState $name raftState
echo "state of aergo3 = $raftState"

if [ "$raftState" = "ProgressStateSnapshot" ]; then
	echo "=========== fail : state must not be snapshot =========="
	exit 100
fi

echo "============== success to catch crash of aergo3 =========="


# restart aergo3
kill_svr.sh 11003
run_svr.sh 11003
checkSync 10001 10003 20

