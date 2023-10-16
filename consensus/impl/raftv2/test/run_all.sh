#!/usr/bin/env bash
# raft 관련 모든 test를 실행한다.
# required tool : jq
source set_test_env.sh
source test_common.sh

echo "kill_svr & clean all test data"
clean_test.sh
init_test.sh 

# raft server boot & down test
echo "pushd $TEST_RAFT_INSTANCE"
pushd $TEST_RAFT_INSTANCE

clean.sh all #remove log
rm BP*

export TEST_SKIP_GENESIS=0
make_node.sh
export TEST_SKIP_GENESIS=1
#RUN_TEST_SCRIPT test_tx.sh 100
#RUN_TEST_SCRIPT test_up_down.sh
#RUN_TEST_SCRIPT test_leader_change.sh 10
#RUN_TEST_SCRIPT test_slow_follower.sh
#RUN_TEST_SCRIPT test_slow_leader.sh
#RUN_TEST_SCRIPT test_slow_follower_restart.sh
#RUN_TEST_SCRIPT test_syncer_crash.sh 0
#RUN_TEST_SCRIPT test_syncer_crash.sh 1
#RUN_TEST_SCRIPT test_member.sh
RUN_TEST_SCRIPT test_new_backup.sh
#RUN_TEST_SCRIPT test_join_syncer_crash.sh 1
#RUN_TEST_SCRIPT test_join_backup.sh
#RUN_TEST_SCRIPT test_invalid_member_init.sh
#ean.sh all RUN_TEST_SCRIPT test_invalid_member_add.sh
#RUN_TEST_SCRIPT test_join_best_entry.sh
popd
