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

	httpports[$nodename]=$((11000 + $i))

	if [ -e "$TEST_RAFT_INSTANCE/$svrport.id" ]; then 
		peerids[$nodename]=`cat $TEST_RAFT_INSTANCE/$svrport.id`
		echo "peerids="${peerids[$nodename]}
	fi
done

function RUN_TEST_SCRIPT() {
	shell="/usr/bin/env bash"
	echo "run scripts> $*"

	$shell -c "$*"
	if [ $? != 0 ]; then
		echo "Test Failed"
		exit 1000
	fi

	echo "Test Succeeded"
}


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
		eval "$2=0"
		return
    fi

    local _height_=$(aergocli -p $port blockchain | jq .Height)
#	echo "getHeight _height_=$_height_"

	if [ "$_height_" = "" ]; then
		_height_=0
	fi

    eval "$2=$_height_"
}

function getHash() {
    local port=$1
    local height=$2

	if [ $# != 3 ];then
		echo "usage: getHash port height retHash"
		exit 100
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
	local i=0
	for (( ; ; )); do
		for i in 10001 10002 10003 10004 10005 ; do
			getLeaderOf $i curleader
			if [[ "$curleader" == aergo* ]]; then
				break
			fi
		done

		echo "curleader=$curleader"

		if [[ "$curleader" != aergo* ]]; then
			echo "<get leader failed>"
			eval "$1="
			sleep 3
			continue
		fi

		echo "leader=$curleader"

		eval "$1=$curleader"
		return
	done
}


function getLeaderOf() {
	local _leader_=""
	if [ $# -ne 2 ]; then
		echo "usage: getLeaderOf rpcport retvalue"
		exit 100
	fi

	local myport=$1

	_leader_=$(aergocli -p $myport blockchain | jq .ConsensusInfo.Status.Leader)
	_leader_=${_leader_//\"/}

	if [[ "$_leader_" == aergo* ]]; then
		eval "$2=$_leader_"
	fi
}

function HasLeader() {
	if [ $# -ne 2 ]; then
		echo "usage: HasLeader rpcport retvalue"
		exit 100
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
		exit 100
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
		exit 200
		eval $3=""
	fi

	echo "<raftid=$_raftID>"
	eval "$3=$_raftID"
}

function getRaftState() {
	local name=$1
	
	if [ "$#" != 2 ]; then 
		echo "Usage: getRaftState servername outStateVar"
		exit 100
	fi

	local leaderPort=
	getLeaderPort leaderPort
	if [ $? -ne 0 ];then
		echo "failed to get leader port"
		exit 100
	fi

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
		exit 100
	fi

	local _leader=""
	getleader _leader

	local _leaderport=${ports[$_leader]}

	if [ "$_leaderport" == "" ];then
		echo "failed to get leader port"
		exit 100
	fi

	echo "leader port=$_leaderport"

	eval "$1=$_leaderport"
}

function isStableLeader() {
	if [ $# -ne 1 ]; then
		echo 'Usage: isStableLeader timeout. return value=$?'
		exit 100
	fi

	mydate=$(date)
	echo "$mydate> check if leader is stable"
	timeout=$1

	local _prevleader=""
	local _tmpLeader=""

	getleader _prevleader
	getleader _tmpLeader

	for ((i=1;i<=$timeout;i++))
	do 
		if [ "$_prevleader" != "$_tmpLeader" ]; then
			echo "Fail: leader changed prev=$_prevleader, cur=$_tmpLeader" 
			return 0
		fi

		sleep 1
	done

	return 1
}

function changeLeader() {
	if [ "$#" != 0 ];then
		echo "Usage: changeLeader"
		exit 100
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
		exit 100
	fi

	# 아무노드나 골라서 5초동안 chain이 증가하고 있는지 확인
	local srcPort=$1
	local timeout=$2
	local heightStart=""
	local heightEnd=""

	local tryHangAgain=10

	echo "isChainHang($timeout) from $srcPort"

	for ((i=1;i<=$tryHangAgain;i++))
	do 
		getHeight $srcPort heightStart

		sleep $timeout

		getHeight $srcPort heightEnd

		echo "start:$heightStart ~ end:$heightEnd"
		if [ "$heightEnd" != "$heightStart" ];then
			echo "check succed"
			return 0
		fi

		echo "chain is hanged. and retry[$i]"
		sleep 1
	done

	return 1
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
	local _srcHeight=
	local _curHeight=

	echo "============ checkSync $srcPort vs $curPort . timeout=$3sec ==========="
	echo "src=$srcPort, curPort=$curPort, time=$timeout"

	for ((i = 1; i<= $3; i++)); do
		sleep 1

		_srcHeight=""
		_curHeight=""
		getHeight $srcPort _srcHeight
		
		getHeight $curPort _curHeight

		echo "srcno=$_srcHeight, curno=$_curHeight"

		if [ "$_srcHeight" = "0" ] || [ "$_curHeight" = "0" ] || [ "$_srcHeight" = "255" ] || [ "$_curHeight" = "255" ]; then
			continue
		fi

		targetNo=$((_curHeight + 3))
		if [ $targetNo -gt $_srcHeight ]; then
			echo "sync succeed"
			isChainHang $curPort 3
			echo ""
			echo ""
			hang=$?
			if [ $hang = 1 ];then
				echo "========= hang after sync ============"
				exit 100
			fi
			return
		fi
	done

	echo "========= sync failed ============"
	exit 100
}

function checkSyncWithLeader() {
	local _curPort=$1
	local _timeout=$2

	echo "============ checkSync with Leader $_curPort . timeout=$2sec ==========="

	local leaderport=
	getLeaderPort leaderport
	if [ $? -ne 0 -o "$leaderport" = "" ];then
		echo "failed to get leader port"
		exit 100
	fi

	checkSync $leaderport $_curPort $_timeout
	if [ $? -ne 0 ];then
		echo "failed to sync with leader $leaderport"
		exit 100
	fi
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

        getHeight $curPort curHeight

        curHash=""
        getHash $curPort $curHeight curHash

        srcHash=""
        getHash $srcPort $curHeight srcHash

		mydate=$(date)
        echo "$mydate> curHeight=$curHeight, srchash=$srcHash, curhash=$curHash"

        if [ "$curHeight" = "-1" ] || [ "$curHash" = "-1" ] || [ "$srcHash" = "-1" ]; then
			echo "========= sync failed ============"
			exit 100 
        fi

        if [ "$curHash" != "$srcHash"  ]; then
			echo "========= sync failed ============"
			exit 100
        fi

        sleep 1
	done

	mydate=$(date)
	echo "$mydate =========== sync is running well =========="
	return 0
}


# copy BP1100[N].toml and _genesis.* to $TEST_RAFT_INSTANCE 
function prepareConfig() {
	if [ $# != "1" ];then
		echo "Usage: $0 configMax"
		exit 100
	fi

	if [ "$TEST_RAFT_INSTANCE" = "" ];then
		echo "TEST_RAFT_INSTANCE is not set"
		exit 100
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

function getAdminUnlocked() {
	if [ $# -ne 3 ]; then
		echo "Usage: $0 rpcport genesis.json retAddress"
		exit 100
	fi

	rpcport=$1
	_genesis="$2"

	if [ ! -e  "$_genesis" ];then
		echo "not exit genesis_wallet.txt ($_genesis)"
		exit 100
	fi

	_admin=`cat $_genesis`

	if [ "$_admin" != "$_ret" ];then
		echo "failed to unlock $_admin"
	fi

    eval "$3=$_admin"
}

function makeAddMemberJson() {
	# return valude is printed by echo
	if [ $# -ne 1 ]; then
		#echo "Usage: $0 nodename"
		exit 100
	fi

	_nodename=$1
	if [[ "$_nodename " != aergo* ]]; then
		#echo "Usage: $0 nodename"
		exit 100
	fi

	memberJson='[ { "command": "add", "name": "'$_nodename'", "address": "/ip4/127.0.0.1/tcp/'${httpports[$_nodename]}'", "peerid":"'${peerids[$_nodename]}'" } ]'

	echo $memberJson
}

function makeRemoveMemberJson() {
	# return valude is printed by echo
	if [ $# -ne 1 ]; then
		#echo "Usage: $0 raftID"
		exit 100
	fi

	_raftID=$1
	_memberJson='[ { "command": "remove", "id":"'$_raftID'" } ]'

	echo $_memberJson
}

function getClusterTotal() {
	_chkPort=$1
	_total=`aergocli -p $_chkPort blockchain | jq .ConsensusInfo.Status.Total`
	echo $(printf %d $_total)
}

function waitClusterTotal() {
	if [ $# -lt 2 -o ! $1 -ge 0 ];then
		echo "Usage: waitClusterTotal totalcount tryCount(every 3 second) reqport"
		exit 100
	fi

	reqCount=$1
	tryCnt=$2
	reqPort=$3
	local i
	echo "Wait cluster: reqCount=$1 tryCnt=$2 reqPort=$3"

	local leaderport=$reqPort

	for ((i=1; i<= $tryCnt ; i++)); do
		if [ "$reqport" = "" ];then
			getLeaderPort leaderport

			if [ $? -ne 0 -o "$leaderport" = "" ];then
				echo "failed to get leader port"
				sleep 3
				continue
			fi
		fi

		total=`aergocli -p $leaderport blockchain | jq .ConsensusInfo.Status.Total`
		echo "i=$i, total=$total, req=$reqCount"
		if [ "$total" = "$reqCount" ];then
			return 1	
		fi

		sleep 3 
		echo $i
	done
		

	echo "failed waitClusterTotal($total)"
	return 0
}

function WaitPeerConnect() {
	if [ $# -ne 2 ]; then
		echo "Usage:$0 expectPeerCount Timeout(sec)"
		exit 100
	fi 

	_reqcnt=$1
	_timeout=$2
	local _res

	for ((i = 1; i<= $_timeout; i++)); do
		_res=$(aergocli -p 10001 getpeers | jq ".|length")

		if [[ "$_res" == fail* ]]; then
			continue
		fi

        echo "WaitPeerConnect: $_res, $_reqcnt "
		if [ "$_res" != "" -a $_res -gt 0 -a $_res -ge $_reqcnt ];then
			echo "peer[$_res] connected"
			return 1
		fi

		sleep 1
	done

	echo "failed peer connection: peer req=$_reqcnt, res=$_res connected."
	return 0
}


function WaitShutdown() {
	if [ $# -ne 2 ]; then
		echo "Usage:$0 pattern Timeout(sec)"
		exit 100
	fi 

	_pattern=$1
	_timeout=$2
	local _res

	for ((i = 1; i<= $_timeout; i++)); do
		existProcess $_pattern
		if [ "$?" == "0" ]; then
			echo "no process $_pattern"
			return 1
		fi

		sleep 1
	done

	echo "failed to wait shutdown(pattern=$_pattern)"
	return 0
}
