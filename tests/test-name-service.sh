#!/usr/bin/env bash

assert_equals() {
  local var="$1"
  local expected="$2"

  if [[ ! "$var" == "$expected" ]]; then
    echo "Assertion failed: $var != $expected"
    echo "File: \"$0\", Line: \"$3\""
    exit 1
  fi
}



../bin/aergocli account import --keystore . --if 47zh1byk8MqWkQo5y8dvbrex99ZMdgZqfydar7w2QQgQqc7YrmFsBuMeF1uHWa5TwA1ZwQ7V6 --password bmttest


echo "-- deploy contract --"

../bin/aergoluac --payload test-name-service.lua > test.out

txhash=$(../bin/aergocli --keystore . \
        contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  --payload `cat test.out` --password bmttest | jq .hash | sed 's/"//g')

sleep 1

sc_id=`../bin/aergocli receipt get $txhash | jq '.contractAddress' | sed 's/"//g'`



echo "-- call contract with an invalid address --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnX"]' \
	--password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_equals "$ret"    "[Contract.LuaResolve] Data and checksum don't match"


echo "-- call contract with a valid address --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnE"]' \
	--password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "AmgExqUu6J4Za8VjyWMJANxoRaUvwgngGQJgemHgwWvuRSEd3wnE"


echo "-- call contract with invalid name --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["long_name-with-invalid.chars"]' \
	--password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    ""


echo "-- call contract with valid but not set name --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["testnametest"]'   --password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    ""


echo "-- register a new account name --"

txhash=$(../bin/aergocli --keystore . name new --name="testnametest" \
	--from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	--password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')

assert_equals "$status" "SUCCESS"


echo "-- call contract --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["testnametest"]'   --password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R"


echo "-- transfer the name --"

txhash=$(../bin/aergocli --keystore . name update --name="testnametest" \
	--from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	--to Amh9vfP5My5DpSafe3gcZ1u8DiZNuqHSN2oAWehZW1kgB3XP4kPi \
	--password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json

status=$(cat receipt.json | jq .status | sed 's/"//g')

assert_equals "$status" "SUCCESS"


echo "-- call contract --"

txhash=$(../bin/aergocli --keystore . contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
	${sc_id} resolve '["testnametest"]'   --password bmttest | jq .hash | sed 's/"//g')

sleep 1

../bin/aergocli receipt get $txhash > receipt.json
status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "Amh9vfP5My5DpSafe3gcZ1u8DiZNuqHSN2oAWehZW1kgB3XP4kPi"


echo ""
echo "All tests pass"
