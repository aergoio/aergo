#!/usr/bin/env bash
source set_test_env.sh

pushd $TEST_RAFT_INSTANCE

while [ 1 ]; do 
	for file in BP*.toml; do
		echo "file=$file"
		port=$(grep 'netserviceport' $file | awk '{ print $3 }')
		aergocli -p $port blockchain
	done
sleep 2
	echo "---------------"
done

popd
