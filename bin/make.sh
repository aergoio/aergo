#!/bin/bash 



set -e 
if [ "$#" -ne 3 ]; then
	    echo "./make.sh [nr_acc] [nr_txs] [target_dir]"
	    exit 
fi

NR_ACC=$1
NR_TX=$2
TARGET_DIR=$3
TO_ACC=ApvPDY2fu55E6ztva5BCZ1FwxZtHvq56ikBcoCUQbedBQSakFdaZ
rm -fr $TARGET_DIR
mkdir -p $TARGET_DIR


echo "Make $NR_ACC account(/w $NR_TX transactions) in $TARGET_DIR..."
for ((i = 1; i <= $NR_ACC; i++))
do
	NEWKEY=`./aergocli newaccount --password 1 2> /dev/null`
	echo $NEWKEY
	./aergocli unlockaccount --address $NEWKEY --password 1 &> /dev/null
	echo "[" > "$TARGET_DIR/$NEWKEY.trx"
	for ((j = 1; j <= $NR_TX; j++))
	do
		printf '\rGenerating..... %d/%d account (%d/%d)' $i $NR_ACC $j $NR_TX
		./aergocli signtx --jsontx \
			"{\"account\":\"$NEWKEY\", \
			\"nonce\": $j , \
			\"price\": 1 , \
			\"limit\": 100 , \
			\"recipient\":\"$TO_ACC\", \
			\"type\": 0, \
			\"amount\": 25000 }" --address $NEWKEY --password 1 >> "$TARGET_DIR/$NEWKEY.trx" 2> /dev/null

		echo "," >> $TARGET_DIR/$NEWKEY.trx
	done
	truncate -s -2 $TARGET_DIR/$NEWKEY.trx
	echo "]" >> $TARGET_DIR/$NEWKEY.trx

done

echo ""


