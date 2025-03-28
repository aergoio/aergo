# stop on errors
set -e
source common.sh

arg=$1
if [ "$arg" != "sbp" ] && [ "$arg" != "dpos" ] && [ "$arg" != "raft" ] && [ "$arg" != "brick" ]; then
  echo "Usage: $0 [brick|sbp|dpos|raft]"
  exit 1
fi
echo "Running integration tests for $arg"

if [ "$arg" == "brick" ]; then
  # run the brick test
  ../bin/brick -V test.brick
  ../bin/brick pack test-import-1.lua
  # compare the output with the expected output. it will fail if the output is different.
  diff test-import-1-bundle.lua test-import-expected-bundle.lua
  exit 0
fi

consensus=$arg

if [ "$consensus" == "sbp" ]; then
  # delete and recreate the aergo folder
  rm -rf ./aergo-files
  mkdir aergo-files
  # copy the config file
  cp config-sbp.toml ./aergo-files/config.toml
  # delete the old logs
  rm -f logs
else
  # delete and recreate the aergo folder
  rm -rf node1
  rm -rf node2
  rm -rf node3
  mkdir node1
  mkdir node2
  mkdir node3
  # copy the config files
  cp config-node1.toml node1/config.toml
  cp config-node2.toml node2/config.toml
  cp config-node3.toml node3/config.toml
  # delete the old logs
  rm -f logs1 logs2 logs3
  # create the genesis block
  echo "creating genesis block..."
  ../bin/aergosvr init --genesis ./genesis-$consensus.json --home ./node1
  ../bin/aergosvr init --genesis ./genesis-$consensus.json --home ./node2
  ../bin/aergosvr init --genesis ./genesis-$consensus.json --home ./node3
fi

# define the config files according to the consensus
if [ "$consensus" == "sbp" ]; then
  config_files=("./aergo-files/config.toml")
elif [ "$consensus" == "dpos" ]; then
  config_files=("./node1/config.toml" "./node2/config.toml" "./node3/config.toml")
elif [ "$consensus" == "raft" ]; then
  config_files=("./node1/config.toml" "./node2/config.toml" "./node3/config.toml")
fi

# define which port used for queries
if [ "$consensus" == "sbp" ]; then
  query_port="7845"
else
  query_port="9845"
fi

echo ""
echo "starting nodes..."
start_nodes
# wait the node(s) to be ready, expecting hardfork version 2
wait_version 2

function set_version() {
  # stop on errors
  set -e
  version=$1
  echo ""
  echo "---------------------------------"
  echo "now test hardfork version $version"
  # get the current block number / height
  block_no=$(../bin/aergocli blockchain | jq .height | sed 's/"//g')
  # increment 2 numbers
  block_no=$((block_no+2))
  # terminate the server process(es)
  stop_nodes
  # save the hardfork config on the config file(s)
  echo "updating the config file(s)..."
  for config_file in "${config_files[@]}"; do
    sed -i "s/^v$version = \"10000\"$/v$version = \"${block_no}\"/" $config_file
  done
  # restart the aergo nodes
  echo "restarting the aergo nodes..."
  start_nodes
  # wait the node(s) to be ready, expecting the new hardfork version
  wait_version $version
  echo "---------------------------------"
  # do not stop on errors
  set +e
}

# do not stop on errors
set +e

num_failed_tests=0

function check() {
    name=$(basename -s .sh $1)
    echo ""
    echo "RUN: $name"
    $@ $version
    local status=$?
    if [ $status -ne 0 ]; then
        num_failed_tests=$((num_failed_tests+1))
        echo "FAIL: $name"
    else
        echo "OK: $name"
    fi
}

# make these variables accessible to the called scripts
export consensus
export query_port

# create the account used on tests
echo "creating user account..."
../bin/aergocli account import --keystore . --if 47zh1byk8MqWkQo5y8dvbrex99ZMdgZqfydar7w2QQgQqc7YrmFsBuMeF1uHWa5TwA1ZwQ7V6 --password bmttest

# run the integration tests - version 2
check ./test-gas-deploy.sh
check ./test-gas-op.sh
check ./test-gas-bf.sh
check ./test-gas-verify-proof.sh
check ./test-gas-per-function-v2.sh
check ./test-contract-deploy.sh
check ./test-pcall-events.sh
check ./test-transaction-types.sh
check ./test-name-service.sh

# change the hardfork version
set_version 3

# run the integration tests - version 3
check ./test-max-call-depth.sh
check ./test-gas-deploy.sh
check ./test-gas-op.sh
check ./test-gas-bf.sh
check ./test-gas-verify-proof.sh
check ./test-gas-per-function-v3.sh
check ./test-contract-deploy.sh
check ./test-pcall-events.sh
check ./test-transaction-types.sh
check ./test-name-service.sh

# change the hardfork version
set_version 4

# run the integration tests - version 4
check ./test-max-call-depth.sh
check ./test-gas-deploy.sh
check ./test-gas-op.sh
check ./test-gas-bf.sh
check ./test-gas-verify-proof.sh
check ./test-gas-per-function-v4.sh
check ./test-contract-deploy.sh
check ./test-pcall-events.sh
check ./test-transaction-types.sh
check ./test-name-service.sh
check ./test-multicall.sh
check ./test-disabled-functions.sh

# terminate the server process
echo ""
echo "closing the aergo nodes"
echo ""
stop_nodes

# print the summary
if [ $num_failed_tests -gt 0 ]; then
  echo "$num_failed_tests failed tests"
  exit 1
else
  echo "All tests pass!"
fi
