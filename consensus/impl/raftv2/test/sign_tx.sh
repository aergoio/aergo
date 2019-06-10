#!/bin/bash
# 각 지갑에서 1 aer씩 송금하는 tx를 tx count만큼 생성한다.

# 어떤 클라이언트
port=$1
TARGET_DIR=$1
# 지갑 갯수
accCount=$2
# tx 수량
txCount=$3

echo "============== sign txs for all accounts =============="

#aergocli -p ${port} account unlock --address ${wallet} --password ${password}

# 계정 & 트랜잭션 삭제
rm -rf ./account_${port}.txt
#rm -rf result.txt
rm -rf ./${port}
#mkdir $TARGET_DIR


chain_id_hash=`aergocli  -p ${port} blockchain | jq .ChainIdHash`
echo "chainid=$chain_id_hash"

# 계정 추가
for ((i = 1; i <= $accCount; i++))
do
TARGET_DIR=${port}/${i}
account=`aergocli account new --password 1234 --path ${TARGET_DIR}`
echo "make tx for account=$account in $TARGET_DIR"
make_tx_internal.sh $TARGET_DIR $account $txCount $chain_id_hash

echo $account >> ./account_${port}.txt
done
