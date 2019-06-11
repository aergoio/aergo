#!/usr/bin/env bash
source test_common.sh

addnode=$1

if [ "$1" = "" ] || [[ "$1" != aergo* ]] ;then
	echo "use:add_member.sh aergo4|aergo5|aergo6|aergo7 usebackup"
	exit 1
fi

usebackup=0
if [ "$2" == "usebackup" ]; then
	usebackup=1	

	echo "try to join $addnode with backup"
fi

declare -A ports svrports svrname httpports peerids

leader=$(aergocli -p 10001 blockchain | jq .ConsensusInfo.Status.Leader)
leader=${leader//\"/}
if [[ $leader != aergo* ]]; then
	echo "leader not exist"
	exit 1
fi

leaderport=${ports[$leader]}

echo "leader=$leader, port=$leaderport"

echo "aergocli -p $leaderport cluster add --name \"$addnode\" --url \"http://127.0.0.1:${httpports[$addnode]}\" --peerid \"${peerids[$addnode]}\""

aergocli -p $leaderport cluster add --name "$addnode" --url "http://127.0.0.1:${httpports[$addnode]}" --peerid "${peerids[$addnode]}"

echo "add Done" 

mySvrport=${svrports[$addnode]}
mySvrName=${svrname[$addnode]}
myConfig="$mySvrName.toml"


for i in {1..5}; do
	echo ${svrname["aergo$i"]}
done
echo "new svr=$mySvrport $mySvrName, $myConfig"

sleep 5

echo "Get config of added member"
echo "cp $TEST_RAFT_INSTANCE_CONF/$myConfig $TEST_RAFT_INSTANCE"
cp $TEST_RAFT_INSTANCE_CONF/$myConfig $TEST_RAFT_INSTANCE

if [ "$usebackup" == "0" ]; then
	echo "init genesis for $mySvrName"
	init_genesis.sh $mySvrName > /dev/null 2>&1
else 
	echo "join using backup: $mySvrName"
	do_sed.sh $myConfig "joinusingbackup=false" "joinusingbackup=true" "/"
	run_svr.sh $mySvrName
fi