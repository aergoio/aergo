set -e
source common.sh


# deploy 66 identical contracts using test-max-call-depth-2.lua
# and store the returned addresses in an array

echo "-- deploy 66 contracts --"

declare -a txhashes
declare -a addresses

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')

../bin/aergoluac --payload test-max-call-depth-2.lua > payload.out

for i in {1..66}
do
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

  txhashes[$i]=$txhash
done

for i in {1..66}
do
  get_receipt ${txhashes[$i]}

  status=$(cat receipt.json | jq .status | sed 's/"//g')
  address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

  assert_equals "$status" "CREATED"
  addresses[$i]=$address
done

# build a list of contract IDs in JSON format

echo "-- build list of contract IDs --"

json="["
for i in {1..66}
do
  address=${addresses[$i]}
  json="$json\"$address\""
  if [ $i -lt 66 ]; then
    json="$json,"
  fi
done
json="$json]"

# call the first contract with a depth of 64

echo "-- call contract with depth 64 --"

#echo "call ${addresses[1]} call_me [$json,1,64]"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,64]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "64"

# check state on all the 64 contracts

echo "-- check state on all the 64 contracts --"

for i in {1..64}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "1"

  ../bin/aergocli contract query ${addresses[$i]} get_call_info '[1]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "$i"
done

# call the first contract with a depth of 66

echo "-- call contract with depth 66 --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,66]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_contains  "$ret" "exceeded the maximum call depth"

# check state on all the 64 contracts

echo "-- check state on all the 64 contracts --"

for i in {1..64}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "1"

  ../bin/aergocli contract query ${addresses[$i]} get_call_info '[1]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "$i"
done

# check state on the 66th contract

echo "-- check state on the 66th contract --"

../bin/aergocli contract query ${addresses[66]} get_total_calls '[]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"

../bin/aergocli contract query ${addresses[66]} get_call_info '[1]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"



# deploy 66 identical contracts using test-max-call-depth-2.lua
# and store the returned addresses in an array

echo "-- deploy 66 contracts --"

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')

../bin/aergoluac --payload test-max-call-depth-3.lua > payload.out

declare -a txhashes
declare -a addresses

for i in {1..66}
do
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

  txhashes[$i]=$txhash
done

for i in {1..66}
do
  get_receipt ${txhashes[$i]}

  status=$(cat receipt.json | jq .status | sed 's/"//g')
  address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

  assert_equals "$status" "CREATED"
  addresses[$i]=$address
done

# get the last nonce for this account

echo "-- get the last nonce for this account --"

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')

# set next_contract for each contract

echo "-- set next_contract for each contract --"

for i in {1..65}
do
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    ${addresses[$i]} set_next_contract "[\"${addresses[$i+1]}\"]" | jq .hash | sed 's/"//g')
  txhashes[$i]=$txhash
done

for i in {1..65}
do
  get_receipt ${txhashes[$i]}

  status=$(cat receipt.json | jq .status | sed 's/"//g')

  assert_equals "$status" "SUCCESS"
done

# call the first contract with a depth of 64

echo "-- call contract with depth 64 --"

#echo "call ${addresses[1]} call_me [1,64]"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[1,64]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "64"

# check state on all the 64 contracts

echo "-- check state on all the 64 contracts --"

for i in {1..64}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "1"

  ../bin/aergocli contract query ${addresses[$i]} get_call_info '[1]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "$i"
done


# call the first contract with a depth of 66

echo "-- call contract with depth 66 --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[1,66]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_contains  "$ret" "exceeded the maximum call depth"

# check state on all the 64 contracts

echo "-- check state on all the 64 contracts --"

for i in {1..64}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "1"

  ../bin/aergocli contract query ${addresses[$i]} get_call_info '[1]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "$i"
done

# check state on the 66th contract

echo "-- check state on the 66th contract --"

../bin/aergocli contract query ${addresses[66]} get_total_calls '[]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"

../bin/aergocli contract query ${addresses[66]} get_call_info '[1]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"



# Circle: contract 1 calls contract 2, contract 2 calls contract 3, contract 3 calls contract 1...

echo "=== Circle ==="

# deploy 4 identical contracts using test-max-call-depth-2.lua

echo "-- deploy 4 contracts --"

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')

../bin/aergoluac --payload test-max-call-depth-2.lua > payload.out

declare -a txhashes
declare -a addresses

for i in {1..4}
do
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

  txhashes[$i]=$txhash
done

for i in {1..4}
do
  get_receipt ${txhashes[$i]}

  status=$(cat receipt.json | jq .status | sed 's/"//g')
  address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

  assert_equals "$status" "CREATED"
  addresses[$i]=$address
done

# build a list of contract IDs in JSON format

echo "-- build list of contract IDs --"

json="["
for i in {1..4}
do
  address=${addresses[$i]}
  json="$json\"$address\""
  if [ $i -lt 4 ]; then
    json="$json,"
  fi
done
json="$json]"

# call the first contract with a depth of 64

echo "-- call contract with depth 64 --"

#echo "call ${addresses[1]} call_me [$json,1,64]"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,64]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "64"

# check state on all the 4 contracts
# each contract should have (64 / 4) = 16 calls

echo "-- check state on all the 4 contracts --"

for i in {1..4}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "16"

  for j in {1..16}
  do
    ../bin/aergocli contract query ${addresses[$i]} get_call_info "[$j]" > receipt.json
    result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
    assert_equals "$result" "$((i+4*(j-1)))"
  done
done


# call the first contract with a depth of 66

echo "-- call contract with depth 66 --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,66]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_contains  "$ret" "exceeded the maximum call depth"

# check state on all the 4 contracts
# each contract should have (64 / 4) = 16 calls

echo "-- check state on all the 4 contracts --"

for i in {1..4}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "16"

  for j in {1..16}
  do
    ../bin/aergocli contract query ${addresses[$i]} get_call_info "[$j]" > receipt.json
    result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
    assert_equals "$result" "$((i+4*(j-1)))"
  done
done

# check state on the 66th contract

echo "-- check state on the 66th contract --"

../bin/aergocli contract query ${addresses[66]} get_total_calls '[]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"

../bin/aergocli contract query ${addresses[66]} get_call_info '[1]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"



# ZigZag: contract 1 calls contract 2, contract 2 calls contract 1...

echo "=== ZigZag ==="

# deploy 2 identical contracts using test-max-call-depth-2.lua

echo "-- deploy 2 contracts --"

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')

../bin/aergoluac --payload test-max-call-depth-2.lua > payload.out

declare -a txhashes
declare -a addresses

for i in {1..2}
do
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

  txhashes[$i]=$txhash
done

for i in {1..2}
do
  get_receipt ${txhashes[$i]}

  status=$(cat receipt.json | jq .status | sed 's/"//g')
  address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

  assert_equals "$status" "CREATED"
  addresses[$i]=$address
done

# build a list of contract IDs in JSON format

echo "-- build list of contract IDs --"

json="["
for i in {1..2}
do
  address=${addresses[$i]}
  json="$json\"$address\""
  if [ $i -lt 2 ]; then
    json="$json,"
  fi
done
json="$json]"

# call the first contract with a depth of 64

echo "-- call contract with depth 64 --"

#echo "call ${addresses[1]} call_me [$json,1,64]"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,64]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    "64"

# check state on all the 2 contracts
# each contract should have (64 / 2) = 32 calls

echo "-- check state on all the 2 contracts --"

for i in {1..2}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "32"

  for j in {1..32}
  do
    ../bin/aergocli contract query ${addresses[$i]} get_call_info "[$j]" > receipt.json
    result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
    assert_equals "$result" "$((i+2*(j-1)))"
  done
done


# call the first contract with a depth of 66

echo "-- call contract with depth 66 --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  ${addresses[1]} call_me "[$json,1,66]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_contains  "$ret" "exceeded the maximum call depth"

# check state on all the 2 contracts
# each contract should have (64 / 2) = 32 calls

echo "-- check state on all the 2 contracts --"

for i in {1..2}
do
  ../bin/aergocli contract query ${addresses[$i]} get_total_calls '[]' > receipt.json
  result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
  assert_equals "$result" "32"

  for j in {1..32}
  do
    ../bin/aergocli contract query ${addresses[$i]} get_call_info "[$j]" > receipt.json
    result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
    assert_equals "$result" "$((i+2*(j-1)))"
  done
done

# check state on the 66th contract

echo "-- check state on the 66th contract --"

../bin/aergocli contract query ${addresses[66]} get_total_calls '[]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"

../bin/aergocli contract query ${addresses[66]} get_call_info '[1]' > receipt.json
result=$(cat receipt.json | sed 's/"//g' | sed 's/\\//g' | sed 's/ //g' | sed 's/value://g')
assert_equals "$result" "null"
