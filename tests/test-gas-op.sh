set -e
source common.sh

fork_version=$1


echo "-- deploy --"

deploy ../contract/vm_dummy/test_files/gas_op.lua

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

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      ""

if [ "$fork_version" -eq "4" ]; then
  assert_equals "$gasUsed"  "120832"
else
  assert_equals "$gasUsed"  "117610"
fi
