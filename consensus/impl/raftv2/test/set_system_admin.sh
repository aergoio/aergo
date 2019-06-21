#!/usr/bin/env bash
source set_test_env.sh
source test_common.sh

pushd $TEST_RAFT_INSTANCE

if [ ! -e genesis -o ! -e genesis.json -o -e wif.tx ]; then
	echo "Err: Not exist genesis account(required files: genesis, genesis.json, wif.txt)"
	exit 100
fi

_leaderport_=
getLeaderPort _leaderport_
if [ $? -ne 0 -o "$_leaderport_" = "" ];then
	echo "failed to get leader port"
	exit 100
fi

# set admin
CLI="aergocli -p $_leaderport_"
ADMIN=
getAdminUnlocked $_leaderport_ ./genesis_wallet.txt ADMIN

echo `$CLI account unlock --address $ADMIN --password 1234`
$CLI account unlock --address $ADMIN --password 1234
echo "$CLI contract call --governance $ADMIN aergo.enterprise appendAdmin '[\"$ADMIN\"]'"
$CLI contract call --governance $ADMIN aergo.enterprise appendAdmin '["'$ADMIN'"]'

popd
sleep 5
