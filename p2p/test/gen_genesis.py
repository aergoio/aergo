#!/usr/bin/python3

# generate config and genesis files for test
import json
import os
import string
import subprocess

import conf
import conf_type
import sys
from builtins import len, int, dict, open, set

nodeCount = len(conf.nodes)
# pkgen 명령 위치(path)를 적어줄 것
cliCmd = "./aergocli"


def low_bool(b):
    if b:
        return "true"
    else:
        return "false"


def generate_genesis(conf, template):
    genesis_filename = "genesis_%s.json" % conf.testname

    bps = list()
    for n in conf.nodes:
        if n.role == conf_type.Role.producer:
            bps.append(n.peerid)

    gaergo_surfix = '000000000000000000'
    total_balance = 500000000
    unit_balance =   10000000
    remain = total_balance
    holders = {}

    i = 0
    first_holder = conf.nodes[conf.holders[0]]
    for idx in conf.holders[1:]:
        n = conf.nodes[idx]
        servId = "s%02d" % idx
        addrProc = subprocess.run(["cat", "keys/%s.addr" % servId], capture_output=True)
        addrProc.check_returncode()
        address = addrProc.stdout.decode("utf-8")
        holders[address] = str(unit_balance) + gaergo_surfix
        remain -= unit_balance
    servId = "s%02d" % conf.holders[0]
    addrProc = subprocess.run(["cat", "keys/%s.addr" % servId], capture_output=True)
    addrProc.check_returncode()
    address = addrProc.stdout.decode("utf-8")
    holders[address] = str(remain) + gaergo_surfix

    d = dict()
    d['magic'] = conf.gen_magic
    d['bps'] = conf_type.wrap_list_to_json(bps)
    d['holders'] = conf_type.wrap_map_to_json(holders)

    genesis_raw = template.substitute(d)
    genesis_formatted = json.dumps(genesis_raw, indent=4)
    conf_file = open(genesis_filename, "w")
    conf_file.write(genesis_raw)
    conf_file.close()

if __name__ == '__main__':
    cwd = os.getcwd()

    argc = len(sys.argv)
    if argc == 2:
        startingPort = int(sys.argv[1])
    else:
        print("Usage: %s <startingPort> " % sys.argv[0])
        sys.exit(1)

    conf_type.setup_nodes(conf.nodes, startingPort)

    tmplfile = open("genesis_json.template", 'r')
    tmplString = tmplfile.read()
    tmplfile.close()
    template = string.Template(tmplString)

    generate_genesis(conf, template)
    print("Finished!")
