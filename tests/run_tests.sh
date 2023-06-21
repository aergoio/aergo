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
