#!/usr/bin/env bash

source set_test_env.sh
source test_common.sh

kill_svr.sh

if [ "$TEST_SKIP_GENESIS" = "1" ];then
	echo "================ skip init genesis node and reboot aergosvr ============="
	if [ ! -e $TEST_RAFT_INSTANCE/BP11001.toml ];then
		echo "not exist BP11001.toml"
		exit 100
	fi

	run_svr.sh

	WaitPeerConnect 2 60
	if [ $? -ne 1 ];then
		echo "failed to connect peer of $file for 60 sec. "
		exit 100
	fi
	exit 0
else
	prepareConfig 3
fi

pushd $TEST_RAFT_INSTANCE

clean.sh
rm init_*.log


if [ $# != 0 ]; then
    echo "Usage: $0"
    exit 100
fi


rm -rf genesis
rm -f genesis.json

for file in BP*.toml; do
    bpname=${file%%.toml}
    echo "./init_genesis.sh $bpname"
#init_genesis.sh $bpname > /dev/null 2>&1
    init_genesis.sh $bpname 
done

if [ "$TEST_NOWAIT_PEER" != "1" ];then
	WaitPeerConnect 2 60
	if [ $? -ne 1 ];then
		echo "failed to connect peer of $file for 60 sec. "
		exit 100
	fi
fi

popd
