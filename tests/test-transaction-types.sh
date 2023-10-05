set -e
source common.sh

fork_version=$1


# contract without "default" function
cat > test-tx-type-1.lua << EOF
function default()
  return system.getAmount()
end

function default2()
  return system.getAmount()
end

abi.payable(default2)
EOF

# contract with not payable "default" function
cat > test-tx-type-2.lua << EOF
function default()
  return system.getAmount()
end

abi.register(default)
EOF

# contract with payable "default" function
cat > test-tx-type-3.lua << EOF
function default()
  return system.getAmount()
end

abi.payable(default)
EOF

echo "-- deploy 1 --"
deploy test-tx-type-1.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
no_default=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

echo "-- deploy 2 --"
deploy test-tx-type-2.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
not_payable=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

echo "-- deploy 3 --"
deploy test-tx-type-3.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
payable_default=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

# delete the contract files
rm test-tx-type-*.lua


# get some info
from=AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R
chainIdHash=$(../bin/aergocli blockchain | jq -r '.ChainIdHash')


echo "-- TRANSFER type, contract without 'default' function --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from $from --to $no_default --amount 1aergo \
  | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "ERROR"
assert_equals "$ret"      "'default' is not payable"
#assert_equals "$gasUsed"  "117861"


echo "-- TRANSFER type, contract with not payable 'default' function --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from $from --to $not_payable --amount 1aergo \
  | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "ERROR"
assert_equals "$ret"      "'default' is not payable"
#assert_equals "$gasUsed"  "117861"


echo "-- TRANSFER type, contract with payable 'default' function --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from $from --to $payable_default --amount 1aergo \
  | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
assert_equals "$ret"      "1000000000000000000"
#assert_equals "$gasUsed"  "117861"


nonce=$(../bin/aergocli getstate --address $from | jq -r '.nonce')

#echo "-- TRANSFER type, trying to make a call --"


echo "-- NORMAL type, contract without 'default' function --"

nonce=$((nonce + 1))

#"Payload": "'$from'",

jsontx='{
"Account": "'$from'",
"Recipient": "'$no_default'", 
"Amount": "1.23aergo",
"Type": 0,
"Nonce": '$nonce',
"chainIdHash": "'$chainIdHash'"}'

jsontx=$(../bin/aergocli --keystore . --password bmttest \
  signtx --address $from --jsontx "$jsontx" )

txhash=$(../bin/aergocli committx --jsontx "$jsontx" | jq .results[0].hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$status"   "ERROR"
	assert_equals "$ret"      "tx not allowed recipient"
	#assert_equals "$gasUsed"  "117861"
else
	assert_equals "$status"   "ERROR"
	assert_equals "$ret"      "'default' is not payable"
	#assert_equals "$gasUsed"  "117861"
fi


echo "-- NORMAL type, contract with not payable 'default' function --"

nonce=$((nonce + 1))

jsontx='{
"Account": "'$from'",
"Recipient": "'$not_payable'",
"Amount": "1.23aergo",
"Type": 0,
"Nonce": '$nonce',
"chainIdHash": "'$chainIdHash'"}'

jsontx=$(../bin/aergocli --keystore . --password bmttest \
  signtx --address $from --jsontx "$jsontx" )

txhash=$(../bin/aergocli committx --jsontx "$jsontx" | jq .results[0].hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$status"   "ERROR"
	assert_equals "$ret"      "tx not allowed recipient"
	#assert_equals "$gasUsed"  "117861"
else
	assert_equals "$status"   "ERROR"
	assert_equals "$ret"      "'default' is not payable"
	#assert_equals "$gasUsed"  "117861"
fi


echo "-- NORMAL type, contract with payable 'default' function --"

nonce=$((nonce + 1))

jsontx='{
"Account": "'$from'",
"Recipient": "'$payable_default'",
"Amount": "1.23aergo",
"Type": 0,
"Nonce": '$nonce',
"chainIdHash": "'$chainIdHash'"}'

jsontx=$(../bin/aergocli --keystore . --password bmttest \
  signtx --address $from --jsontx "$jsontx" )

txhash=$(../bin/aergocli committx --jsontx "$jsontx" | jq .results[0].hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
	assert_equals "$status"   "ERROR"
	assert_equals "$ret"      "tx not allowed recipient"
	#assert_equals "$gasUsed"  "117861"
else
	assert_equals "$status"   "SUCCESS"
	assert_equals "$ret"      "1230000000000000000"
	#assert_equals "$gasUsed"  "117861"
fi



echo "-- CALL type, contract without 'default' function (not sending) --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $no_default default | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "ERROR"
assert_equals "$ret"      "undefined function: default"
#assert_equals "$gasUsed"  "117861"


echo "-- CALL type, contract without 'default' function (sending)  --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $no_default default --amount 1aergo | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "ERROR"
assert_equals "$ret"      "'default' is not payable"
#assert_equals "$gasUsed"  "117861"


echo "-- CALL type, contract with not payable 'default' function (not sending) --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $not_payable default | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
assert_equals "$ret"      "0"
#assert_equals "$gasUsed"  "117861"


echo "-- CALL type, contract with not payable 'default' function (sending) --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $not_payable default --amount 1aergo | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "ERROR"
assert_equals "$ret"      "'default' is not payable"
#assert_equals "$gasUsed"  "117861"


echo "-- CALL type, contract with payable 'default' function --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $payable_default default --amount 1aergo | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
assert_equals "$ret"      "1000000000000000000"
#assert_equals "$gasUsed"  "117861"
