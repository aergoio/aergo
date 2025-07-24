set -e
source common.sh

fork_version=$1

# Test network commands
echo "--- Testing network commands ---"

# Test getpeers
echo -n "Testing getpeers... "
peers=$(../bin/aergocli getpeers)
if [ -z "$peers" ]; then
  echo "ERROR: Failed to get peers"
  exit 1
fi
echo "✓"

# Test getconsensusinfo
echo -n "Testing getconsensusinfo... "
consensus_info=$(../bin/aergocli getconsensusinfo)
if [ -z "$consensus_info" ] || ! echo "$consensus_info" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get consensus info"
  exit 1
fi
echo "✓"

# Test system commands
echo "--- Testing system commands ---"

# Test serverinfo
echo -n "Testing serverinfo... "
server_info=$(../bin/aergocli serverinfo)
if [ -z "$server_info" ] || ! echo "$server_info" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get server info"
  exit 1
fi
echo "✓"

# Test chaininfo
echo -n "Testing chaininfo... "
chain_info=$(../bin/aergocli chaininfo)
if [ -z "$chain_info" ] || ! echo "$chain_info" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get chain info"
  exit 1
fi
echo "✓"

# Test chainstat
echo -n "Testing chainstat... "
chain_stat=$(../bin/aergocli chainstat)
if [ -z "$chain_stat" ] || ! echo "$chain_stat" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get chain stats"
  exit 1
fi
echo "✓"

# Test utility commands
echo "--- Testing utility commands ---"

# Test version
echo -n "Testing version... "
version=$(../bin/aergocli version)
if [ -z "$version" ]; then
  echo "ERROR: Failed to get version"
  exit 1
fi
echo "✓"

# Test keygen
echo -n "Testing keygen... "
key_info=$(../bin/aergocli keygen --password bmttest)
if [ -z "$key_info" ]; then
  echo "ERROR: Failed to generate key"
  exit 1
fi
echo "✓"

# Test metric
echo -n "Testing metric... "
metric_info=$(../bin/aergocli metric)
if [ -z "$metric_info" ]; then
  echo "ERROR: Failed to get metrics"
  exit 1
fi
echo "✓"

# Test account commands
echo "--- Testing account commands ---"

# Test account new
echo -n "Testing account new... "
account=$(../bin/aergocli account new --keystore . --password bmttest | grep -o "Am.*")
if [ -z "$account" ]; then
  echo "ERROR: Failed to create new account"
  exit 1
fi
echo "✓"

# Test account list
echo -n "Testing account list... "
accounts=$(../bin/aergocli account list --keystore .)
if [[ ! "$accounts" == *"$account"* ]]; then
  echo "ERROR: Newly created account not found in account list"
  exit 1
fi
echo "✓"

# Test account import (using existing test data)
echo -n "Testing account import... "
../bin/aergocli account import --keystore . --password bmttest --if 47zh1byk8MqWkQo5y8dvbrex99ZMdgZqfydar7w2QQgQqc7YrmFsBuMeF1uHWa5TwA1ZwQ7V6 >/dev/null 2>&1
echo "✓"

# Test blockchain commands
echo "--- Testing blockchain commands ---"

# Test blockchain info
echo -n "Testing blockchain info... "
chain_info=$(../bin/aergocli blockchain)
if [ -z "$chain_info" ] || ! echo "$chain_info" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get blockchain info"
  exit 1
fi
echo "✓"

# Test getblock
echo -n "Testing getblock... "
block=$(../bin/aergocli getblock --number 1)
if [ -z "$block" ] || ! echo "$block" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get block 1"
  exit 1
fi
echo "✓"

# Test getstate
echo -n "Testing getstate... "
state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
if [ -z "$state" ] || ! echo "$state" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get state for test account"
  exit 1
fi
echo "✓"

# Test transaction commands
echo "--- Testing transaction commands ---"

# Test sendtx
echo -n "Testing sendtx... "
txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  --to AmQ2tS46doLfqsH4vy7PXmvGn7paec8wgqPWmGrcJKiLZADrJcME \
  --amount 1aergo | jq .hash | sed 's/"//g')

if [ -z "$txhash" ] || [ ${#txhash} -lt 43 ]; then
  echo "ERROR: Failed to send transaction"
  exit 1
fi
echo "✓           txn hash: $txhash"

# Test gettx
echo -n "Testing gettx... "
sleep 1
transaction=$(../bin/aergocli gettx $txhash)
if [ -z "$transaction" ] || ! echo "$transaction" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get transaction"
  exit 1
fi
echo "✓"

# Wait for receipt
get_receipt $txhash

# Test receipt
echo -n "Testing receipt... "
receipt=$(../bin/aergocli receipt get $txhash)
if [ -z "$receipt" ] || ! echo "$receipt" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get receipt"
  exit 1
fi
echo "✓"

# Test contract commands
echo "--- Testing contract commands ---"

# Test contract deploy
echo -n "Testing contract deploy... "
deploy ../contract/vm_dummy/test_files/all_types.lua
if [ -z "$txhash" ] || [ ${#txhash} -lt 43 ]; then
  echo "ERROR: Failed to deploy all_types.lua contract"
  exit 1
fi
echo "✓  txn hash: $txhash"

# Wait for receipt
echo -n "Waiting for receipt... "
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/\"//g')
contract_address=$(cat receipt.json | jq .contractAddress | sed 's/\"//g')
if [ "$status" != "CREATED" ]; then
  echo "ERROR: contract deployment failed with status: $status"
  exit 1
fi
echo "✓   deployed to: $contract_address"

# Test contract abi
echo -n "Testing contract abi... "
abi=$(../bin/aergocli contract abi $contract_address)
if [ -z "$abi" ] || ! echo "$abi" | jq . >/dev/null 2>&1; then
  echo "ERROR: Failed to get contract ABI"
  exit 1
fi
echo "✓"

# Test contract calls (3 different types)
echo "--- Testing contract calls ---"

# Test value_set call
echo -n "Testing value_set call... "
txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $contract_address value_set '["testname"]' | jq .hash | sed 's/\"//g')

if [ -z "$txhash" ] || [ ${#txhash} -lt 43 ]; then
  echo "ERROR: Failed to call value_set function"
  exit 1
fi
echo "✓   txn hash: $txhash"

echo -n "Waiting for receipt... "
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/\"//g')
if [ "$status" != "SUCCESS" ]; then
  echo "ERROR: value_set function call failed with status: $status"
  exit 1
fi
echo "✓"

# Test array_set call
echo -n "Testing array_set call... "
txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $contract_address array_append '["testarray"]' | jq .hash | sed 's/\"//g')

if [ -z "$txhash" ] || [ ${#txhash} -lt 43 ]; then
  echo "ERROR: Failed to call array_append function"
  exit 1
fi
echo "✓   txn hash: $txhash"

echo -n "Waiting for receipt... "
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/\"//g')
if [ "$status" != "SUCCESS" ]; then
  echo "ERROR: array_append function call failed with status: $status"
  exit 1
fi
echo "✓"

# Test map_set call
echo -n "Testing map_set call... "
txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $contract_address map_set '["testkey", "testvalue"]' | jq .hash | sed 's/\"//g')

if [ -z "$txhash" ] || [ ${#txhash} -lt 43 ]; then
  echo "ERROR: Failed to call map_set function"
  exit 1
fi
echo "✓     txn hash: $txhash"

echo -n "Waiting for receipt... "
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/\"//g')
if [ "$status" != "SUCCESS" ]; then
  echo "ERROR: map_set function call failed with status: $status"
  exit 1
fi
echo "✓"

# Test contract queries (3 different types)
echo "--- Testing contract queries ---"

# Test value_get query
echo -n "Testing value_get query... "
../bin/aergocli contract query $contract_address value_get '[]' > query_result.txt 2>&1
query_exit_code=$?
if [ $query_exit_code -ne 0 ]; then
  echo "ERROR: value_get query failed with exit code $query_exit_code."
  cat query_result.txt
  exit 1
else
  query_result=$(cat query_result.txt)
  if [ -z "$query_result" ]; then
    echo "ERROR: value_get query returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$query_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testname"
    echo "✓"
  fi
fi

# Test array_get query
echo -n "Testing array_get query... "
../bin/aergocli contract query $contract_address array_get '[1]' > query_result.txt 2>&1
query_exit_code=$?
if [ $query_exit_code -ne 0 ]; then
  echo "ERROR: array_get query failed with exit code $query_exit_code."
  cat query_result.txt
  exit 1
else
  query_result=$(cat query_result.txt)
  if [ -z "$query_result" ]; then
    echo "ERROR: array_get query returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$query_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testarray"
    echo "✓"
  fi
fi

# Test map_get query
echo -n "Testing map_get query... "
../bin/aergocli contract query $contract_address map_get '["testkey"]' > query_result.txt 2>&1
query_exit_code=$?
if [ $query_exit_code -ne 0 ]; then
  echo "ERROR: map_get query failed with exit code $query_exit_code."
  cat query_result.txt
  exit 1
else
  query_result=$(cat query_result.txt)
  if [ -z "$query_result" ]; then
    echo "ERROR: map_get query returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$query_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testvalue"
    echo "✓"
  fi
fi

# Test contract state queries (3 different state variables)
echo "--- Testing contract state queries ---"

# Test statequery for name (single value)
echo -n "Testing statequery for name... "
../bin/aergocli contract statequery $contract_address name > state_result.txt 2>&1
state_exit_code=$?
if [ $state_exit_code -ne 0 ]; then
  echo "ERROR: State query for name failed with exit code $state_exit_code."
  cat state_result.txt
  exit 1
else
  state_result=$(cat state_result.txt)
  if [ -z "$state_result" ]; then
    echo "ERROR: State query for name returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$state_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testname"
    echo "✓"
  fi
fi

# Test statequery for list (array)
echo -n "Testing statequery for list... "
../bin/aergocli contract statequery $contract_address list 1 > state_result.txt 2>&1
state_exit_code=$?
if [ $state_exit_code -ne 0 ]; then
  echo "ERROR: State query for list failed with exit code $state_exit_code."
  cat state_result.txt
  exit 1
else
  state_result=$(cat state_result.txt)
  if [ -z "$state_result" ]; then
    echo "ERROR: State query for list returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$state_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testarray"
    echo "✓"
  fi
fi

# Test statequery for values (map)
echo -n "Testing statequery for values... "
../bin/aergocli contract statequery $contract_address values testkey > state_result.txt 2>&1
state_exit_code=$?
if [ $state_exit_code -ne 0 ]; then
  echo "ERROR: State query for values failed with exit code $state_exit_code."
  cat state_result.txt
  exit 1
else
  state_result=$(cat state_result.txt)
  if [ -z "$state_result" ]; then
    echo "ERROR: State query for values returned empty result"
    exit 1
  else
    cleaned_result=$(echo "$state_result" | sed 's/"//g' | sed 's/\\//g' | sed 's/value://g')
    assert_equals "$cleaned_result" "testvalue"
    echo "✓"
  fi
fi

# Test name service commands
echo "--- Testing name service commands ---"

# use a different account name for each hardfork
account_name="testuserver$fork_version"

# Test name new
echo -n "Testing name new... "
name_txhash=$(../bin/aergocli --keystore . --password bmttest \
  name new --from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  --name $account_name | jq .hash | sed 's/"//g')

if [ -z "$name_txhash" ] || [ ${#name_txhash} -lt 43 ]; then
  echo "ERROR: Failed to register name"
  exit 1
fi
echo "✓        txn hash: $name_txhash"

# Wait for receipt
echo -n "Waiting for receipt... "
get_receipt $name_txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
if [ "$status" != "SUCCESS" ]; then
  echo "ERROR: name registration failed with status: $status"
  exit 1
fi
echo "✓"
