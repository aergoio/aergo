start_nodes() {

  if [ "$consensus" == "sbp" ]; then
    # open the aergo node in testmode
    ../bin/aergosvr --testmode --home ./aergo-files >> logs 2>> logs &
    pid=$!
  else
    # open the 5 nodes
    ../bin/aergosvr --home ./node1 >> logs1 2>> logs1 &
    pid1=$!
    ../bin/aergosvr --home ./node2 >> logs2 2>> logs2 &
    pid2=$!
    ../bin/aergosvr --home ./node3 >> logs3 2>> logs3 &
    pid3=$!
    # nodes 4 and 5 use a previous version of aergosvr for backwards compatibility check
    docker run --name aergo-test-node4 --rm --net=host -v $(pwd)/node4:/aergo aergo/node:2.8.1 aergosvr --home /aergo >> logs4 2>> logs4 &
    docker run --name aergo-test-node5 --rm --net=host -v $(pwd)/node5:/aergo aergo/node:2.8.1 aergosvr --home /aergo >> logs5 2>> logs5 &
  fi

}

stop_nodes() {

  if [ "$consensus" == "sbp" ]; then
    kill $pid
    # wait until the node is stopped
    wait $pid
  else
    # stop directly executed nodes
    kill $pid1 $pid2 $pid3
    # wait until nodes are stopped
    wait $pid1 $pid2 $pid3
    # stop Docker containers (it will wait until they are stopped)
    docker stop aergo-test-node4 aergo-test-node5 >/dev/null 2>/dev/null || true
  fi

}

wait_version_from() {
  local expect_version=$1
  local port=$2
  local counter=0
  local output
  local cur_version

  while true; do
    # do not stop on errors
    set +e
    # get the current hardfork version
    output=$(../bin/aergocli blockchain --port $port 2>/dev/null)
    # stop on errors
    set -e
    # check if 'output' is non-empty and starts with '{'
    if [[ -n "$output" ]] && [[ "${output:0:1}" == "{" ]]; then
      cur_version=$(echo "$output" | jq .chainInfo.id.version | sed 's/"//g')
      if [ "$cur_version" == "null" ]; then
        cur_version=0
      fi
    else
      cur_version=0
    fi

    #echo "current version: $cur_version"

    if [ $cur_version -eq $expect_version ]; then
      version=$expect_version
      break
    else
      sleep 0.5
      counter=$((counter+1))
      if [ $counter -gt 20 ]; then
        echo "Failed to change the blockchain version on the node at port $port"
        echo "Desired: $expect_version, Actual: $cur_version"
        exit 1
      fi
    fi
  done

}

wait_version() {
  local expect_version=$1

  if [ "$consensus" == "sbp" ]; then
    wait_version_from $expect_version 7845
  else
    wait_version_from $expect_version 7845
    wait_version_from $expect_version 8845
    wait_version_from $expect_version 9845
    wait_version_from $expect_version 10845
    wait_version_from $expect_version 11845
  fi
}

get_deploy_args() {
  local contract_file=$1

  if [ "$fork_version" -ge "4" ]; then
    deploy_args="$contract_file"
  else
    ../bin/aergoluac --payload $contract_file > payload.out
    deploy_args="--payload `cat payload.out`"
  fi

}

deploy() {
  get_deploy_args $1

  # do not stop on errors
  set +e
  # deploy the contract
  ../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args > output.json
  exit_code=$?
  # stop on errors
  set -e

  if [ $exit_code -ne 0 ]; then
    echo "Error: aergocli command failed"
    echo "Command output:"
    cat output.json
    exit 1
  fi

  # check if the JSON response contains an error field
  error_value=$(cat output.json | jq '.error')
  if [ "$error_value" != "null" ]; then
    echo "Error: Transaction failed"
    echo "Command output:"
    cat output.json
    exit 1
  fi

  txhash=$(cat output.json | jq .hash | sed 's/"//g')
}

get_receipt_from() {
  local txhash=$1
  local port=$2
  local receipt_file=$3
  local counter=0

  # wait for a total of (0.4 * 100) = 40 seconds

  while true; do
    # do not stop on errors
    set +e
    # request the receipt for the transaction
    ../bin/aergocli receipt get --port $port $txhash > "$receipt_file" 2> error.txt
    exit_code=$?
    output=$(cat error.txt 2>/dev/null)
    # stop on errors
    set -e

    #echo "output: $output"

    if [[ $exit_code -ne 0 ]] && [[ $output == *"tx not found"* ]]; then
      sleep 0.4
      counter=$((counter+1))
      if [ $counter -gt 100 ]; then
        echo "Error getting receipt on port $port: tx not found: $txhash"
        exit 1
      fi
    elif [[ $exit_code -ne 0 ]]; then
      echo "Error getting receipt on port $port: $output"
      exit 1
    elif ! jq . "$receipt_file" >/dev/null 2>&1; then
      # if output is not valid JSON, wait and retry
      sleep 0.4
      counter=$((counter+1))
      if [ $counter -gt 100 ]; then
        echo "Error getting receipt on port $port: Invalid JSON response: $txhash"
        cat "$receipt_file"
        exit 1
      fi
    else
      # valid JSON response received
      break
    fi
  done

}

get_receipt() {
  local txhash=$1
  local i   # this is VERY IMPORTANT! to avoid overwriting the global variable i
  local retry
  local max_retries=10
  local all_match

  # check txhash length (must be 43 or 44 characters)
  local txlen=${#txhash}
  if [[ $txlen != 43 && $txlen != 44 ]]; then
    echo "Error: Invalid transaction hash: $txhash"
    exit 1
  fi

  if [ "$consensus" == "sbp" ]; then
    get_receipt_from $txhash 7845 receipt.json
    return
  fi

  for retry in $(seq 1 $max_retries); do
    # get receipts from all 5 nodes
    get_receipt_from $txhash 7845 receipt1.json
    get_receipt_from $txhash 8845 receipt2.json
    get_receipt_from $txhash 9845 receipt3.json
    get_receipt_from $txhash 10845 receipt4.json
    get_receipt_from $txhash 11845 receipt5.json

    # compare receipts - they must be exactly the same
    all_match=true
    for i in {2..5}; do
      local diff_output=$(diff receipt1.json receipt${i}.json)
      if [ -n "$diff_output" ]; then
        all_match=false
        echo "Warning: receipt from node $i differs from receipt from node 1"
        break
      fi
    done

    if [ "$all_match" == "true" ]; then
      break
    elif [ $retry -lt $max_retries ]; then
      sleep 3
      echo "Checking consensus again..."
    fi
  done

  if [ "$all_match" != "true" ]; then
    echo "Error: receipts still differ after $max_retries attempts"
    exit 1
  fi

  # rename receipt1.json to receipt.json and delete the others
  mv receipt1.json receipt.json
  rm receipt{2..5}.json
}

get_internal_operations() {
  txhash=$1
  # do not stop on errors
  set +e

  output=$(../bin/aergocli operations $txhash --port $query_port 2>&1 > internal_operations.json)

  #echo "output: $output"

  if [[ $output == *"No internal operations found for this transaction"* ]]; then
    echo -n "" > internal_operations.json
  elif [[ -n $output ]]; then
    echo "Error getting internal operations: $output"
    exit 1
  fi

  # stop on errors
  set -e
}

assert_equals() {
  local var="$1"
  local expected="$2"

  if [[ ! "$var" == "$expected" ]]; then
    echo "Assertion failed: $var != $expected"
    exit 1
  fi
}

assert_contains() {
  local var="$1"
  local substring="$2"

  if [[ ! "$var" == *"$substring"* ]]; then
    echo "Assertion failed: $var does not contain $substring"
    exit 1
  fi
}
