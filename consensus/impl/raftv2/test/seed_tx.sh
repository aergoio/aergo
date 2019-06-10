#!/bin/bash

if [ $# != 3 ];then
	echo "Usage: $0 port genesis_addr startNonce"
fi

# 어떤 클라이언트
port=$1
# 지갑 갯수
genesis_addr=$2
# tx 수량
#count2=$3
j=$3

target=10001

echo ""
echo "================== sign tx ====================="
echo " run : $0 $port $genesis_addr $j"
echo "================================================"

# 지갑 생성 & 언락
aergocli -p ${port} account unlock --address ${genesis_addr} --password 1234

chain_id_hash=`aergocli  -p ${port} blockchain | jq .ChainIdHash`
echo "chainid=$chain_id_hash"

# 트랜잭션 생성

rm -rf $genesis_addr.tmp
touch $genesis_addr.tmp
echo "[" >> $genesis_addr.tmp

# 각 계좌에 종자돈을 송금해 놓는다.
echo "============== sign seed money tx for all accounts =============="
while read line
do
	aergocli -p ${port} signtx --jsontx \
				 "{\"account\":\"$genesis_addr\", \
				 \"nonce\": $j , \
				 \"price\": \"1\" , \
				 \"limit\": 100 , \
				 \"recipient\":\"$line\", \
				 \"type\": 0, \
				 \"chainidhash\": $chain_id_hash, \
				 \"amount\": \"1000000000000000000000000\" }"  \
				 --address $genesis_addr --password 1234 >> $genesis_addr.tmp
	echo "," >>  $genesis_addr.tmp
	echo $j $line
	j=$(($j+1))
done < account_${target}.txt

truncate -s -2 $genesis_addr.tmp
echo "]" >> $genesis_addr.tmp

echo "============== confirm seed money for all accounts =============="
# 트랜잭션 컨펌
aergocli -p ${port} committx --jsontxpath $genesis_addr.tmp
