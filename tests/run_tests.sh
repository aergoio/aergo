#!/usr/bin/env bash

# delete the aergo folder
rm -r ~/.aergo/

# open the aergo server in testmode to create the config file
../bin/aergosvr --testmode > logs 2> logs &
pid=$!
# wait it create the config file
sleep 3
# terminate the server process
kill $pid
# enable the block production on the config file
sed -i 's/^enablebp = false$/enablebp = true/' ~/.aergo/config.toml
# restart the aergo server in testmode
../bin/aergosvr --testmode > logs 2> logs &
pid=$!
sleep 3

# run the integration tests
./test-name-service.sh

# terminate the server process
kill $pid
