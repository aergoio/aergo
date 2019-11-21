#!/bin/bash

# binary 위치
import os
import pathlib
import subprocess
import sys
from builtins import *

import conftype
import conf

svrCmd = "./aergosvr"

def init_datadir(basedir, nodes):
    dataDir = pathlib.Path(cwd+"/data")
    if not dataDir.is_dir():
        os.mkdir(dataDir)

    for i in range(0, len(conf.nodes)):
        n = conf.nodes[i]
        sid = n.rpcPort%100
        datadir = "data/X%05d"%n.rpcPort
        confFile = "X%05d.toml"%n.rpcPort
        outfile = "glog_s%02d.log"

        print("Creating genesis chain data %s"%datadir)

        idProc = subprocess.run([svrCmd, "init", "--config", confFile, "--genesis", genesis_filename, "--home=."])
        idProc.check_returncode()

if __name__ == '__main__':
    cwd = os.getcwd()
    argc = len(sys.argv)
    if argc != 3:
        print ("Usage: $0 <starting port> <genesisfile>")
        sys.exit(1)

    startingPort = int(sys.argv[1])
    genesis_filename = sys.argv[2]
    gen_file = pathlib.Path(genesis_filename)
    if not gen_file.is_file():
        print ("Genesis file %s is not found" % genesis_filename)
        sys.exit(1)
    conftype.setup_nodes(conf.nodes, startingPort)

    init_datadir(cwd+"/data", conf.nodes)
