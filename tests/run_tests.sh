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
echo "starting the aergo server..."
../bin/aergosvr --testmode --home ./aergo-files > logs 2> logs &
pid=$!
# wait it to be ready
sleep 2

version=$(../bin/aergocli blockchain | jq .ChainInfo.Chainid.Version | sed 's/"//g')

function set_version() {
  # stop on errors
  set -e
  version=$1
  echo ""
  echo "--- now tests using version $version ---"
  # get the current block number / height
  block_no=$(../bin/aergocli blockchain | jq .Height | sed 's/"//g')
  # increment 2 numbers
  block_no=$((block_no+2))
  # terminate the server process
  kill $pid
  # save the hardfork config on the config file
  echo "updating the config file..."
  if [ $version -eq 2 ]; then
    sed -i "s/^v2 = \"10000\"$/v2 = \"${block_no}\"/" ./aergo-files/config.toml
  elif [ $version -eq 3 ]; then
    sed -i "s/^v3 = \"10000\"$/v3 = \"${block_no}\"/" ./aergo-files/config.toml
  fi
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

# run the integration tests - version 2
check ./test-gas-deploy.sh
check ./test-gas-op.sh
check ./test-gas-bf.sh
check ./test-gas-verify-proof.sh

# change the hardfork version
set_version 3

# run the integration tests - version 3
check ./test-max-call-depth.sh
check ./test-gas-deploy.sh
check ./test-gas-op.sh
check ./test-gas-bf.sh
check ./test-gas-verify-proof.sh

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
