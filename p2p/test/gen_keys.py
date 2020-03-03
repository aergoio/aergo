# generate config and genesis files for test

import os
import pathlib
import subprocess
from builtins import len

import conf

nodeCount = len(conf.nodes)
# pkgen 명령 위치(path)를 적어줄 것
cliCmd = "./aergocli"


def low_bool(b):
    if b:
        return "true"
    else:
        return "false"


def generate_pks(nodes):
    i = 0
    for n in nodes:
        servId = "s%02d" % i
        genProc = subprocess.run([cliCmd, "keygen", "--addr", servId])
        genProc.check_returncode()
        os.system("mv %s.* keys/" % servId)
        i += 1

if __name__ == '__main__':
    cwd = os.getcwd()

    keyDir = pathlib.Path(cwd + "/keys")
    if not keyDir.is_dir():
        os.mkdir(keyDir)
    generate_pks(conf.nodes)

    print("Finished!")
