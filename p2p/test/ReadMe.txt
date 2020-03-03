This directory contains files and docs for testing p2p features

1. bp-agent work test : this test requires at least three different machines (or different ehternet interfaces)


How to Use
NOTE: some features are not implemented yet.

1. copy script files and template files to appropriate directory
2. create conf.py to configure test settings. (actually, copy conf.py.example and modify content)
3. run gen_keys.py, gen_conf.py, gen_genesis.py and init_svr.py to generate keys, config file, genesis json file
  and initial chain datas. 'python script.py arg1 arg2
4. run dist_files.py to copy files to local test directory, test container or test machine.
5. run test
 