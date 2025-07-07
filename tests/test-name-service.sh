set -e
source common.sh

fork_version=$1


echo "-- deploy contract --"

deploy test-name-service.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- call contract with an invalid address --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnX"]' \
	--password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals   "$status" "ERROR"
  # assert_equals "$ret"    "[Contract.LuaResolve] Data and checksum don't match"
  assert_contains "$ret"    "Data and checksum don't match"
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


echo "-- call contract with a valid address --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnE"]' \
	--password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals "$status" "SUCCESS"
  assert_equals "$ret"    "AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnE"
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


echo "-- call contract with invalid name --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["long_name-with-invalid.chars"]' \
	--password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals "$status" "SUCCESS"
  assert_equals "$ret"    ""
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


echo "-- call contract with valid but not set name --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["testnametest"]'   --password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals "$status" "SUCCESS"
  assert_equals "$ret"    ""
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


# use a different account name for each hardfork
account_name="testnamever$fork_version"
# later, it could also:
#  - drop the account name, to recreate it later
#  - let the account name to expire, by forwarding time


echo "-- register a new account name --"

txhash=$(../bin/aergocli --keystore . name new --name="$account_name" \
	--from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	--password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')

assert_equals "$status" "SUCCESS"


echo "-- call contract --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["'$account_name'"]'   --password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals "$status" "SUCCESS"
  assert_equals "$ret"    "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R"
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


echo "-- transfer the name --"

txhash=$(../bin/aergocli --keystore . name update --name="$account_name" \
	--from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	--to Amh9vfP5My5DpSafe3gcZ1u8DiZNuqHSN2oAWehZW1kgB3XP4kPi \
	--password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')

assert_equals "$status" "SUCCESS"


echo "-- call contract --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${address} resolve '["'$account_name'"]'   --password bmttest | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

if [ "$fork_version" -ge "4" ]; then
  assert_equals "$status" "SUCCESS"
  assert_equals "$ret"    "Amh9vfP5My5DpSafe3gcZ1u8DiZNuqHSN2oAWehZW1kgB3XP4kPi"
else
  assert_equals   "$status" "ERROR"
  assert_contains "$ret"    "attempt to index global 'name_service'"
fi


echo "-- query the contract --"

../bin/aergocli contract query ${address} resolve '["'$account_name'"]' > result.txt 2> result.txt || true
result=$(cat result.txt)

if [ "$fork_version" -ge "4" ]; then
  result=$(echo $result | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
  assert_equals   "$result"  "Amh9vfP5My5DpSafe3gcZ1u8DiZNuqHSN2oAWehZW1kgB3XP4kPi"
else
  assert_contains "$result"  "Error: failed to query contract"
  assert_contains "$result"  "attempt to index global 'name_service'"
fi
