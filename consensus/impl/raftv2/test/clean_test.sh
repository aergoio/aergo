#!/usr/bin/env sh
source set_test_env.sh

kill_svr.sh 
rm -rf $TEST_RAFT_INSTANCE
