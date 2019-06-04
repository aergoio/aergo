#!/usr/bin/env bash
# generate id, BP11001.toml ~ BP1100[N].toml
# this script must run at most once before testing

source set_test_env.sh
source $TEST_RAFT_HOME/test_common.sh 

echo "################################################################################"
echo " INITIALIZE RAFT TEST "
echo ""
echo ""
echo " TEST_RAFT_HOME=$TEST_RAFT_HOME"
echo " TEST_RAFT_INSTANCE_CONF=$TEST_RAFT_INSTANCE_CONF"
echo "################################################################################"

cd $TEST_RAFT_HOME
echo "rm -rf $TEST_RAFT_INSTANCE"
echo "mkdir -p $TEST_RAFT_INSTANCE"
rm -rf $TEST_RAFT_INSTANCE
mkdir -p $TEST_RAFT_INSTANCE
mkdir -p $TEST_RAFT_INSTANCE_DATA
mkdir -p $TEST_RAFT_INSTANCE_CLIENT

# change datadir in config files
echo "copy config files to temporary instance directory"
echo "cp -rf $TEST_RAFT_CONF $TEST_RAFT_INSTANCE_CONF"
cp -rf $TEST_RAFT_CONF $TEST_RAFT_INSTANCE_CONF

echo "pushd $TEST_RAFT_INSTANCE_CONF"
pushd $TEST_RAFT_INSTANCE_CONF

for i in $(seq 1 7); do 
	mynodename="aergo$i"
	mysvrname=${svrname[$mynodename]}
	svrport=${svrports[$mynodename]}
	echo "$mynodename, $mysvrname, $svrport"

	echo "do_sed.sh $mysvrname _data_ $TEST_RAFT_INSTANCE_DATA/$svrport ="
	do_sed.sh $mysvrname  _data_ "$TEST_RAFT_INSTANCE_DATA/$svrport" "="
done

popd

cp -rf $TEST_RAFT_INSTANCE_CONF/1100*.* $TEST_RAFT_INSTANCE
cp -rf $TEST_RAFT_INSTANCE_CONF/arglog.* $TEST_RAFT_INSTANCE

echo " Initialize Done!"
echo "################################################################################"
