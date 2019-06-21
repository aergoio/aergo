#!/usr/bin/env bash
source test_common.sh

if [ "$1" = "" ] ; then
	echo "use:rm_member.sh aergo1~aergo3"
	exit 100
fi


rmnode=$1


# get leader
myleader=
getleader myleader
echo "myleader=$myleader"


getLeaderPort leaderport
prevCnt=$(getClusterTotal $leaderport)


raftID=""
getRaftID $leaderport $rmnode raftID

# get leader port

echo "leader=$myleader, port=$leaderport, raftId=$raftID"

#echo "aergocli -p $leaderport cluster remove --nodeid $raftID"
#aergocli -p $leaderport cluster remove --nodeid $raftID

walletFile="$TEST_RAFT_INSTANCE/genesis_wallet.txt"
ADMIN=
getAdminUnlocked $leaderport $walletFile ADMIN

rmJson="$(makeRemoveMemberJson $raftID)"

echo "aergocli -p "$leaderport" contract call --governance "$ADMIN" aergo.enterprise changeCluster "$rmJson
aergocli -p $leaderport contract call --governance $ADMIN aergo.enterprise changeCluster "$rmJson"
echo "remove Done" 

# check if total count is decremented
reqCnt=$((prevCnt-1))
echo "reqClusterTotal=$reqCnt"
waitClusterTotal $reqCnt $leaderport 10
if [ $? -ne 1 ]; then
	echo "remove failed"
	exit 100
fi
