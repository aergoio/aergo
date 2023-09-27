set -e
source common.sh


echo "-- deploy ARC1 factory --"

../bin/aergoluac --payload ../contract/vm_dummy/test_files/arc1-factory.lua > payload.out

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
arc1_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- deploy ARC2 factory --"

../bin/aergoluac --payload ../contract/vm_dummy/test_files/arc2-factory.lua > payload.out

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
arc2_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- deploy caller contract --"

../bin/aergoluac --payload ../contract/vm_dummy/test_files/token-deployer.lua > payload.out

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` "[\"$arc1_address\",\"$arc2_address\"]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
deployer_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- deploy 1 ARC1 contract --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $arc1_address new_token '["Test","TST",18,1000,{"burnable":true, "mintable":true, "pausable":true, "blacklist":true, "all_approval":true, "limited_approval":true}]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"
#assert_equals "$gasUsed"  "117861"


echo "-- deploy 1 ARC2 contract --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $arc2_address new_arc2_nft '["Test NFT 1","NFT1",null,{"burnable":true, "mintable":true, "metadata":true, "pausable":true, "blacklist":true, "approval":true, "searchable":true, "non_transferable":true, "recallable":true}]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"
#assert_equals "$gasUsed"  "117861"


echo "-- deploy 1 ARC1 and 1 ARC2 contracts --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $deployer_address deploy_tokens '[1,1]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"
#assert_equals "$gasUsed"  "117861"


echo "-- deploy 2 ARC1 and 2 ARC2 contracts --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $deployer_address deploy_tokens '[2,2]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"
#assert_equals "$gasUsed"  "117861"


echo "-- deploy 3 ARC1 and 3 ARC2 contracts --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $deployer_address deploy_tokens '[3,3]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
#assert_equals "$ret"      "{}"
#assert_equals "$gasUsed"  "117861"
