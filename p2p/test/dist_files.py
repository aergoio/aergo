# distribute generated files such as config, initial data, log settings or etc. to target machines
import os
import subprocess
import sys

import conf
import conf_type

LCP_CMD = 'cp'
SCP_CMD = 'scp'


def is_local(machine):
    return len(machine.connInfo)==0

def exec_remote(m, cmd):
    wrapped_cmd="\"%s ;\""%cmd
    destHost = "%s@%s" % (m.userID, m.connAddr)
    print("Printing... ",["ssh","-p",str(m.connPort),destHost,cmd])
    subprocess.Popen(["ssh","-p",str(m.connPort),destHost,cmd])

def copy_files(src_file, target_dir, m ):
    if is_local(m):
        cp = LCP_CMD
        cparg = "-r"
        baseDest = target_dir
    else:
        cp = SCP_CMD
        cparg = ("-P %d" % m.connPort)
        baseDest = "%s@%s:%s" % (m.userID, m.connAddr, target_dir)

    dest = baseDest + "/"
    genProc = subprocess.run([cp, cparg, "-r", src_file, dest])
    genProc.check_returncode()

def create_dir(dir, m):
    if is_local(m):
        genProc = subprocess.run(["mkdir","-p",dir])
        genProc.check_returncode()
    else:
        cmd = "mkdir -p %s"%dir
        exec_remote(m, cmd)

if __name__ == '__main__':
    cwd = os.getcwd()
    argc = len(sys.argv)
    if argc == 2:
        startingPort = int(sys.argv[1])
        genPK = False
    else:
        print("Usage: python %s <startingport>" % sys.argv[0])
        sys.exit(1)

    conf_type.setup_nodes(conf.nodes, startingPort)

    for m in conf.machines :
        print("# %s"%m)

    for n in conf.nodes:
        m = n.machine
        confFile = ("X%05d.toml" % n.rpcPort)
        basedir = m.baseDir
        datadir = "data/"+("X%05d" % n.rpcPort)
        copy_files(confFile, m.baseDir, m)
        create_dir(basedir+"/data", m)
        copy_files(datadir, basedir+"/"+datadir,m)