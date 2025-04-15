set -e
source common.sh

fork_version=$1


echo "-- deploy --"

if [ "$fork_version" -eq "4" ]; then
  deploy ../contract/vm_dummy/test_files/gas_bf_v4.lua
else
  deploy ../contract/vm_dummy/test_files/gas_bf_v2.lua
fi

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- call --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $address main "[]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

if [ "$consensus" = "raft" ]; then
  assert_equals "$status"   "ERROR"
  assert_equals "$ret"      "exceeded the maximum instruction count"
else
  assert_equals "$status"   "SUCCESS"
  assert_equals "$ret"      ""
fi

if [ "$consensus" = "raft" ]; then
  assert_equals "$gasUsed"  "0"
elif [ "$fork_version" -eq "4" ]; then
  assert_equals "$gasUsed"  "47342481"
elif [ "$fork_version" -eq "3" ]; then
  assert_equals "$gasUsed"  "47456046"
else
  assert_equals "$gasUsed"  "47456244"
fi
