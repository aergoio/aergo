#!/bin/bash
metapath=$1
wallet_addr=$2
tx_cnt=$3
chainid=$4

echo "$0"

touch $metapath/$wallet_addr".tmp"

echo "============== make send txs for $metapath/$wallet_addr =============="

echo "[" >> $metapath/$wallet_addr.tmp
	for ((j = 1; j <= $tx_cnt; j++))
	do
aergocli signtx --path $metapath --jsontx \
	"{\"account\":\"$wallet_addr\", \
       	\"nonce\": $j , \
		\"chainidhash\": ${chainid}, \
       	\"recipient\":\"AmPAUu1LCtKCntGG714dzmRpdcFAWWMjedTTqHR32W63Dd5GauKq\", \
	\"amount\": \"1\" }"  --address $wallet_addr --password 1234 >> $metapath/$wallet_addr.tmp
echo "," >>  $metapath/$wallet_addr.tmp
	 done

		 truncate -s -2 $metapath/$wallet_addr.tmp
echo "]" >> $metapath/$wallet_addr.tmp
