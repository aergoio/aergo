#!/usr/bin/env bash
echo "include test_common.sh"
export INCLUDE_TEST_COMMON="YES"

declare -A nodenames ports svrports svrname httpports peerids

ALLIPS="10001 10002 10003 10004 10005 10006 10007"

for i in {1..7} ; do
	nodename="aergo$i"
	nodenames[$i]=$nodename

	ports[$nodename]=$((10000 + $i))
	svrport=$((11000 + $i))
	svrports[$nodename]=$svrport
	svrname[$nodename]="BP$svrport"

	httpports[$nodename]=$((13000 + $i))

	if [ -e "$svrport.id" ]; then 
		peerids[$nodename]=`cat $svrport.id`
	fi
done


function existProcess() {
    local port=$1

    local proc=$(ps  -ef|grep aergosvr | grep $port | awk '{print $2 }')
    if [ "$proc" = "" ]; then
        return "0"
    fi

    return "1"
}

function getHeight() {
    local port=$1

    local serverport=$(($port + 1000))

	#echo "port=$1, serverport=$serverport"
	echo "port=$1"

    existProcess $serverport
    if [ "$?" = "0" ]; then
		echo "no process $serverport"
		exit 100
    fi

    local height=$(aergocli -p $port blockchain | jq .Height)

	if [ "$height" = "" ]; then
		height=0
	fi

    return $height
}

function getHash() {
    local port=$1
    local height=$2

	if [ $# != 3 ];then
		echo "usage: getHash port height retHash"
		exit
	fi

    local serverport=$(($port + 1000))
	echo "serverport=$serverport"

    existProcess $serverport
    if [ "$?" = "0" ]; then
		"no process $serverport"
		exit 100
    fi

	echo "aergocli -p $port getblock --number $height | jq .Hash"
    local hash=$(aergocli -p $port getblock --number $height | jq .Hash)

    eval "$3=$hash"
}

function getleader() {
	local curleader=""
	for i in 10001 10002 10003 10004 10005 ; do
		getLeaderOf $i curleader
		if [[ "$curleader" == aergo* ]]; then
			break
		fi
	done

	echo "curleader=$curleader"

	if [[ "$curleader" != aergo* ]]; then
		echo "<get leader failed>"
		exit
	fi

	echo "leader=$curleader"

	eval "$1=$curleader"
}


function getLeaderOf() {
	local _leader_=""
	if [ $# -ne 2 ]; then
		echo "usage: getLeaderOf rpcport retvalue"
		exit
	fi

	local myport=$1

	_leader_=$(aergocli -p $myport blockchain | jq .ConsensusInfo.Status.Leader)
	_leader_=${_leader_//\"/}

	echo "curleader=$_leader_"

	if [[ "$_leader_" == aergo* ]]; then
		eval "$2=$_leader_"
	fi
}

function HasLeader() {
	if [ $# -ne 2 ]; then
		echo "usage: HasLeader rpcport retvalue"
		exit
	fi

	local myport=$1
	local myRet
	getLeaderOf $myport myRet

	if [[ "$myRet" == aergo* ]]; then
		echo "leader exist in $myport"
		eval "$2=1"
	else
		echo "leader not exist in $myport"
		eval "$2=0"
	fi
}

function getRaftID() {
	if [ $# != 3 ]; then
		echo "getRafTID leaderport name outRaftID"
		exit 1
	fi

	local _leaderport=$1
	local name=$2
	local pattern=".Bps|.[]|select(.Name==\"$name\")|.RaftID"
	local _raftID=

	echo "aergocli -p $_leaderport getconsensusinfo | jq $pattern"
	_raftID=`aergocli -p $_leaderport getconsensusinfo | jq $pattern`
	ret=$?
	if [ "$_raftID" == "" ]; then
		echo "failed to get raftID for $name"
		exit 2
		eval $3=""
	fi

	echo "<raftid=$_raftID>"
	eval "$3=$_raftID"
}

function getRaftState() {
	local name=$1
	
	if [ "$#" != 2 ]; then 
		echo "Usage: getRaftState servername outStateVar"
	fi

	local leaderPort=
	getLeaderPort leaderPort

	# getRaftID
	getRaftID $leaderPort $name raftID

	# getRaftStatus
	local pattern=".Info.Status.progress[\"$raftID\"].state"

	local _raftState=$(aergocli -p $leaderPort getconsensusinfo | jq $pattern)

	echo "<raftState=$_raftState>"

	eval "$2=$_raftState"
}


function getLeaderPort() {
	if [ $# != 1 ]; then
		echo "Usage: getLeaderPort leaderport"
		exit
	fi

	local _leader=""
	getleader _leader

	local _leaderport=${ports[$_leader]}

	if [ "$_leaderport" == "" ];then
		echo "failed to get leader port"
		return 0
		exit
	fi

	eval "$1=$_leaderport"
}


function testxxx() {
	local x
		getleader x

	echo "testleader=$x"
}

function isStableLeader() {
	if [ $# -ne 1 ]; then
		echo 'Usage: isStableLeader timeout. return value=$?'
		exit
	fi

	timeout=$1

	local _prevleader=""
	local _tmpLeader=""

	getleader _prevleader
	getleader _tmpLeader

	for ((i=1;i<=$timeout;i++))
	do 
		if [ "$_prevleader" != "$_tmpLeader" ]; then
			return 0
		fi

		sleep 1
	done

	return 1
}

function changeLeader() {
	if [ "$#" != 0 ];then
		echo "Usage: changeLeader"
		exit
	fi

	local leaderName

	getleader leaderName

	echo "cur leader: $leaderName"
	
	local leaderPort="" 
	leaderPort=${svrports[$leaderName]}
	echo "leaderport=$leaderPort"

	kill_svr.sh $leaderPort
	sleep 2
	DEBUG_CHAIN_BP_SLEEP=$chainSleep run_svr.sh $leaderPort
	sleep 2

	leaderName=""
	getleader leaderName
	echo "new leader: $leaderName"
}

function isChainHang() {
	# "isChainHang: return 1 if true"
	if [ "$#" != "2" ];then
		echo "Usage: isChainHang targetRpcPort timeout"
	fi

	# 아무노드나 골라서 5초동안 chain이 증가하고 있는지 확인
	local srcPort=$1
	local timeout=$2
	local heightStart=""

	echo "isChainHang($timeout) from $srcPort"
	getHeight $srcPort
	heightStart=$?

	sleep $timeout

	local heightEnd=""
	getHeight $srcPort
	heightEnd=$?

	echo "start:$heightStart ~ end:$heightEnd"

	if [ "$heightEnd" = "$heightStart" ];then
		echo "chain is hanged"
		exit 100
	fi

	echo "check succed"

	return 0
}

function checkReorg() {
	reorgCount=$(egrep 'reorg' ./*.log | wc -l | awk '{print $1}')

	if [ "$reorgCount" != "0" ];then
		echo "failed: reorg occured"
		exit 100
	fi
}


# 모든 노드의 leader가 0이 아닌 valid node를 가르키는지
function checkLeaderValid() {
	local curleader=""
	for i in $ALLIPS ; do
		local _svrport=$(($i + 1000))
		existProcess $_svrport
		if [ "$?" = "0" ]; then
			continue
		fi

		curleader=$(aergocli -p $i blockchain | jq .ConsensusInfo.Status.Leader)
		curleader=${curleader//\"/}
		if [[ "$curleader" != aergo* ]]; then
			echo "failed: leader of $i is $curleader"
			exit 100
		fi
	done
}

function checkSync() {
	local srcPort=$1
	local curPort=$2
	local timeout=$3

	echo "============ checkSync $srcPort vs $curPort . timeout=$3sec ==========="
	echo "src=$srcPort, curPort=$curPort, time=$timeout"

	for ((i = 1; i<= $3; i++)); do
		srcHeight=""
		curHeight=""
		getHeight $srcPort 
		srcHeight=$?
		
		getHeight $curPort
		curHeight=$?

		echo "srcno=$srcHeight, curno=$curHeight"

		if [ "$srcHeight" = "-1" ] || [ "$curHeight" = "-1" ]; then
			continue
		fi

		targetNo=$((curHeight + 3))
		if [ $targetNo -gt $srcHeight ]; then
			echo "sync succeed"
			isChainHang $curPort 3
			echo ""
			echo ""
			hang=$?
			if [ $hang = 1 ];then
				echo "========= hang after sync ============"
				exit 1
			fi
			return
		fi

		sleep 1
	done

	echo "========= sync failed ============"
	exit 
}


# 현재 sync가 정상적으로 진행중인지 검사
# 현재 best가 remote 에 connect되어 있는 지 확인
function checkSyncRunning() {
    local srcPort=$1
    local curPort=$2
    local try=$3

    local srcHash
    local curHash
    local curHeight

    echo "============ checkSyncRunning $srcPort vs $curPort . try=$3 nums ==========="

    for ((i = 1; i<= $try; i++)); do
        curHeight=""

        getHeight $curPort
        curHeight=$?

        echo "curHeight=$curHeight"

        curHash=""
        getHash $curPort $curHeight curHash
        echo "curHash=$curHash"

        srcHash=""
        getHash $srcPort $curHeight srcHash
        echo "srcHash=$srcHash"


        echo "curHeight=$curHeight, srchash=$srcHash, curhash=$curHash"

        if [ "$curHeight" = "-1" ] || [ "$curHash" = "-1" ] || [ "$srcHash" = "-1" ]; then
			echo "========= sync failed ============"
			exit 
        fi

        if [ "$curHash" != "$srcHash"  ]; then
			echo "========= sync failed ============"
			exit 
        fi

        sleep 1
	done

	echo "=========== sync is running well =========="
	return 0
}


# copy BP1100[N].toml and _genesis.* to $TEST_RAFT_INSTANCE 
function prepareConfig() {
	if [ $# != "1" ];then
		echo "Usage: $0 configMax"
		exit
	fi

	if [ "$TEST_RAFT_INSTANCE" = "" ];then
		echo "TEST_RAFT_INSTANCE is not set"
		exit
	fi


	configMax=$1
	echo "prepare config files ($configMax)"

	for i in $(seq 1 $configMax); do
		echo "cp  $TEST_RAFT_INSTANCE_CONF/BP1100$i.toml  $TEST_RAFT_INSTANCE"
		cp  $TEST_RAFT_INSTANCE_CONF/BP1100$i.toml  $TEST_RAFT_INSTANCE
	done
	echo "cp  $TEST_RAFT_INSTANCE_CONF/_genesis.* $TEST_RAFT_INSTANCE"
	cp  $TEST_RAFT_INSTANCE_CONF/_genesis.* $TEST_RAFT_INSTANCE
}
