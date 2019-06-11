#!/usr/bin/env bash
source test_common.sh

if [ "$1" = "" ] ; then
	echo "use:rm_member.sh aergo1~aergo3"
	exit 1
fi


rmnode=$1


# get leader
myleader=
getleader myleader
echo "myleader=$myleader"

getLeaderPort curLeaderPort


raftID=""
getRaftID $curLeaderPort $rmnode raftID

# get leader port

echo "leader=$myleader, port=$curLeaderPort, raftId=$raftID"

echo "aergocli -p $curLeaderPort cluster remove --nodeid $raftID"

aergocli -p $curLeaderPort cluster remove --nodeid $raftID

echo "remove Done" 
