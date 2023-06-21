# stop on errors
set -e

# run the brick test
./test-brick.sh

# delete the aergo folder
rm -rf ~/.aergo/

# open the aergo server in testmode to create the config file
echo "starting the aergo server..."
../bin/aergosvr --testmode > logs 2> logs &
pid=$!
# wait it create the config file
sleep 3
# terminate the server process
kill $pid
# enable the block production on the config file
echo "updating the config file..."
sed -i 's/^enablebp = false$/enablebp = true/' ~/.aergo/config.toml
# restart the aergo server in testmode
echo "restarting the aergo server..."
../bin/aergosvr --testmode > logs 2> logs &
pid=$!
sleep 3

# do not stop on errors
set +e

num_failed_tests=0

function check() {
    name=$(basename -s .sh $1)
    echo ""
    echo "RUN: $name"
    $@
    local status=$?
    if [ $status -ne 0 ]; then
        num_failed_tests=$((num_failed_tests+1))
        echo "FAIL: $name"
    else
        echo "OK: $name"
    fi
}

# run the integration tests
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

if [ $num_failed_tests -gt 0 ]; then
  echo "$num_failed_tests failed tests"
  exit 1
else
  echo "All tests pass!"
fi
