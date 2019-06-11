#!/usr/bin/env bash
# raft 관련 모든 test를 실행한다.
# required tool : jq
source set_test_env.sh

rm -rf $TEST_RAFT_INSTANCE

init_test.sh 

# raft server boot & down test
echo "pushd $TEST_RAFT_INSTANCE"
pushd $TEST_RAFT_INSTANCE

echo "kill_svr & clean"
kill_svr.sh
clean.sh all #remove log
rm BP*

export TEST_SKIP_GENESIS=0
test_tx.sh 100
export TEST_SKIP_GENESIS=1

test_up_down.sh
test_leader_change.sh 10
test_slow_follower.sh
test_slow_leader.sh
test_syncer_crash.sh 0
test_syncer_crash.sh 1
test_member.sh
test_join_backup.sh
test_invalid_member_init.sh
test_invalid_member_add.sh
popd
