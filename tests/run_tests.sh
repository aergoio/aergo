# stop on errors
set -e

# run the brick test
./test-brick.sh

# delete and recreate the aergo folder
rm -rf ./aergo-files
mkdir aergo-files
# copy the config file
cp config.toml ./aergo-files/

# open the aergo server in testmode to create the config file
echo ""
echo "starting the aergo server..."
../bin/aergosvr --testmode --home ./aergo-files > logs 2> logs &
pid=$!
# wait it to be ready
sleep 2

version=$(../bin/aergocli blockchain | jq .ChainInfo.Chainid.Version | sed 's/"//g')
if [ $version -ne 2 ]; then
  echo "Wrong hardfork version!"
  echo "Desired: 2, Actual: $version"
  exit 1
fi

function set_version() {
  # stop on errors
  set -e
  version=$1
  echo ""
  echo "---------------------------------"
  echo "now test hardfork version $version"
  # get the current block number / height
  block_no=$(../bin/aergocli blockchain | jq .Height | sed 's/"//g')
  # increment 2 numbers
  block_no=$((block_no+2))
  # terminate the server process
  kill $pid
  # save the hardfork config on the config file
  echo "updating the config file..."
  sed -i "s/^v${version} = \"10000\"$/v${version} = \"${block_no}\"/" ./aergo-files/config.toml
  # restart the aergo server
  echo "restarting the aergo server..."
  ../bin/aergosvr --testmode --home ./aergo-files > logs 2> logs &
  pid=$!
  # wait it to be ready
  sleep 3
  # check if it worked
  new_version=$(../bin/aergocli blockchain | jq .ChainInfo.Chainid.Version | sed 's/"//g')
  if [ $new_version -ne $version ]; then
    echo "Failed to change the blockchain version on the server"
    echo "Desired: $version, Actual: $new_version"
    exit 1
  fi
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
check ./test-transaction-types.sh

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
check ./test-transaction-types.sh

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
check ./test-transaction-types.sh

# terminate the server process
echo ""
echo "closing the aergo server"
echo ""
kill $pid

# print the summary
if [ $num_failed_tests -gt 0 ]; then
  echo "$num_failed_tests failed tests"
  exit 1
else
  echo "All tests pass!"
fi
