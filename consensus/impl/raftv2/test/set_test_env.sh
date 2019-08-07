#!/usr/bin/env bash

export SET_TEST_ENV="YES"

echo "include set_test_env.sh"

BASEDIR="$( cd "$(dirname "$0")" ; pwd -P )"
export PATH=$(pwd):$PATH

export TEST_RAFT_HOME="$BASEDIR"
export TEST_RAFT_CONF="$TEST_RAFT_HOME/config"
export TEST_RAFT_INSTANCE="$TEST_RAFT_HOME/tmp"
export TEST_RAFT_INSTANCE_DATA="$TEST_RAFT_HOME/tmp/data"
export TEST_RAFT_INSTANCE_CONF="$TEST_RAFT_HOME/tmp/config"
export TEST_RAFT_INSTANCE_CLIENT="$TEST_RAFT_HOME/tmp/client"

if [ ! -d "$TEST_RAFT_INSTANCE_DATA" ];then
	mkdir -p $TEST_RAFT_INSTANCE_DATA
fi

if [ ! -d "$TEST_RAFT_INSTANCE_CONF" ];then
	mkdir -p $TEST_RAFT_INSTANCE_CONF
fi

if [ ! -d "$TEST_RAFT_INSTANCE_CLIENT" ];then
	mkdir -p $TEST_RAFT_INSTANCE_CLIENT
fi

