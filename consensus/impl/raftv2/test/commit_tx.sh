#!/bin/bash

echo "=========== commit tx for all accounts =============="
# 어떤 클라이언트
port=$1
TARGET_DIR=$2

if [ "$port" = "" ];then
   port=10001
fi

if [ "$2" = "" ]; then
	TARGET_DIR=$port
fi

echo "targetdir=$TARGET_DIR"

# 계정 & 트랜잭션 삭제
echo "start" 
aergocli -p ${port} blockchain

# 트랜잭션 컨펌
for file in $TARGET_DIR/**/*.tmp; do
    echo $file " confirm .."
 
	if [ "$file" = ".tmp" ]; then
	 continue;
 	fi

	aergocli -p ${port} committx --jsontxpath ${file}
done

echo "end"
aergocli  -p ${port} blockchain
