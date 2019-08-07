#!/usr/bin/env bash

if [ $# != 2 ];then
	echo "Usage: $0 port txPerAcc"
	exit 100
fi

port=$1
if [ "$port" == "" ]; then
    port=10001
fi

txPerAcc=$2
if [ "$txPerAcc" == "" ]; then
    txPerAcc=1000
fi

accountNum=10

rm -rf ./*.txt ./$port ./*.tmp


echo "================ generate txs acc=$accountNum txs=$txPerAcc ==============="

server_dir=$TEST_RAFT_INSTANCE

genesis_wallet=$(cat $server_dir/genesis_wallet.txt)
echo "my genesis wallet=$genesis_wallet"
if [ "$genesis_wallet" = "" ]; then
	echo "genesis_wallet is empty"
	exit 100
fi


sign_tx.sh ${port} $accountNum $txPerAcc
seed_tx.sh $port $genesis_wallet 1

sleep 5
