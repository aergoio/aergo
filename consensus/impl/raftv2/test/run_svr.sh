#!/usr/bin/env bash
source set_test_env.sh
source test_common.sh

pushd $TEST_RAFT_INSTANCE

BP_NAME=""

#rm BP*.toml
#./aergoconf-gen.sh 10001 tmpl.toml 5
#./make_node.sh  10001 tmpl.toml 1234
if [ -z "$1" ];then
	pattern="BP.*toml"
else
	pattern="$1"
fi

for file in $(ls BP* | grep $pattern); do
	echo $file
	BP_NAME=$(echo $file | cut -f 1 -d'.')
	if [ "${BP_NAME}" != "tmpl" -a "${BP_NAME}" != "arglog" ]; then
		echo "${BP_NAME}: aergosvr --config $file >> server_${BP_NAME}.log 2>&1"
		aergosvr --config $file >> server_${BP_NAME}.log 2>&1 &
	fi
done

sleep 3


popd
