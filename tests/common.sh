
start_nodes() {

  if [ "$consensus" == "sbp" ]; then
    # open the aergo node in testmode
    ../bin/aergosvr --testmode --home ./aergo-files > logs 2> logs &
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
    docker run --name aergo-test-node4 --rm --net=host -v $(pwd)/node4:/aergo aergo/node:2.6.0 aergosvr --home /aergo >> logs4 2>> logs4 &
    docker run --name aergo-test-node5 --rm --net=host -v $(pwd)/node5:/aergo aergo/node:2.6.0 aergosvr --home /aergo >> logs5 2>> logs5 &
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

wait_version() {
  expect_version=$1
  counter=0
  # do not stop on errors
  set +e

  while true; do
    # get the current hardfork version
    output=$(../bin/aergocli blockchain 2>/dev/null)
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
        echo "Failed to change the blockchain version on the nodes"
        echo "Desired: $expect_version, Actual: $cur_version"
        exit 1
      fi
    fi
  done

  # stop on errors
  set -e
}

get_deploy_args() {
  contract_file=$1

  if [ "$fork_version" -ge "4" ]; then
    deploy_args="$contract_file"
  else
    ../bin/aergoluac --payload $contract_file > payload.out
    deploy_args="--payload `cat payload.out`"
  fi

}

deploy() {

  get_deploy_args $1

  txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args | jq .hash | sed 's/"//g')

}

get_receipt_from() {
  txhash=$1
  port=$2
  receipt_file=$3
  counter=0
  # do not stop on errors
  set +e

  # wait for a total of (0.4 * 100) = 40 seconds

  while true; do
    output=$(../bin/aergocli receipt get --port $port $txhash 2>&1 > $receipt_file)

    #echo "output: $output"

    if [[ $output == *"tx not found"* ]]; then
      sleep 0.4
      counter=$((counter+1))
      if [ $counter -gt 100 ]; then
        echo "Error: tx not found: $txhash"
        exit 1
      fi
    elif [[ -n $output ]]; then
      echo "Error: $output"
      exit 1
    else
      break
    fi
  done

  # stop on errors
  set -e
}

get_receipt() {
  txhash=$1

  if [ "$consensus" == "sbp" ]; then
    get_receipt_from $txhash 7845 receipt.json
    return
  fi

  # Get receipts from all 5 nodes
  get_receipt_from $txhash 7845 receipt1.json
  get_receipt_from $txhash 8845 receipt2.json
  get_receipt_from $txhash 9845 receipt3.json
  get_receipt_from $txhash 10845 receipt4.json
  get_receipt_from $txhash 11845 receipt5.json

  # Compare receipts - they must be exactly the same
  for i in {2..5}; do
    diff_output=$(diff receipt1.json receipt${i}.json)
    if [ -n "$diff_output" ]; then
      echo "Error: Receipt from node 1 differs from receipt from node $i"
      exit 1
    fi
  done

  # Rename receipt1.json to receipt.json and delete the others
  mv receipt1.json receipt.json
  rm receipt{2..5}.json
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
