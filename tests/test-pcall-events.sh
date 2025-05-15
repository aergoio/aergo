set -e
source common.sh

fork_version=$1


echo "-- deploy --"

deploy ../contract/vm_dummy/test_files/pcall-events-3.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address3=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


get_deploy_args ../contract/vm_dummy/test_files/pcall-events-2.lua

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args '["'$address3'"]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address2=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


get_deploy_args ../contract/vm_dummy/test_files/pcall-events-1.lua

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args '["'$address2'"]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address1=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


get_deploy_args ../contract/vm_dummy/test_files/pcall-events-0.lua

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args '["'$address2'"]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address0=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- pcall --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $address1 test_pcall "[]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')
nevents=$(cat receipt.json | jq '.events | length')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$nevents" "2"
else
	assert_equals "$nevents" "6"
fi


echo "-- xpcall --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $address1 test_xpcall "[]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')
nevents=$(cat receipt.json | jq '.events | length')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$nevents" "2"
else
	assert_equals "$nevents" "6"
fi


echo "-- contract.pcall --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $address1 test_contract_pcall "[]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')
nevents=$(cat receipt.json | jq '.events | length')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$nevents" "2"
else
	assert_equals "$nevents" "6"
fi


#echo "----------- contract-1 event list ------------"
#aergocli event list --address $address1 --recent 1000

#echo "----------- contract-2 event list ------------"
#aergocli event list --address $address2 --recent 1000

#echo "----------- contract-3 event list ------------"
#aergocli event list --address $address3 --recent 1000
