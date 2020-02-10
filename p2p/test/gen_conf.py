#!/usr/bin/python3

# generate config and genesis files for test

import os
import pathlib
import string
import subprocess
import sys
from builtins import len, int, dict, open, set

import conf
import conf_type

from conf_type import wrap_list_to_json, low_bool

nodeCount = len(conf.nodes)
# pkgen 명령 위치(path)를 적어줄 것
cliCmd = "./aergocli"


def generate_pks(nodes):
    i = 1
    for n in nodes:
        servId = "s%02d" % i
        genProc = subprocess.run([cliCmd, "keygen", servId])
        genProc.check_returncode()
        os.system("mv %s.* keys/" % servId)
        i += 1


def generate_confs(nodes, template):
    i = 0
    for n in nodes:
        servId = "s%02d" % i
        confFileName = "X%05d.toml" % n.rpcPort
        d = dict()
        d['home'] = cwd
        d['hostaddr'] = n.machine.ip
        d['rpc'] = n.rpcPort
        d['p2pport'] = n.p2pPort
        d['profport'] = n.profPort
        d['pkfilename'] = "%s.key" % servId
        d['enablebp'] = low_bool(n.role == 'p')
        d['expose'] = low_bool(not n.hidden)
        d['discover'] = low_bool(not n.hidden)
        d['usepolaris'] = low_bool(not n.hidden)
        if not n.hidden :
            d['polarises'] = wrap_list_to_json(conf.polarises)
        else:
            d['polarises'] = '[]'
        d['role'] = n.role.name.lower()
        if n.role == conf_type.Role.producer:
            d['agent'] = n.agent
        else:
            d['agent'] = ""
        d['producers'] = wrap_list_to_json(n.producers)

        to_add = set_toadd(i, n)
        peers = []
        for idx in to_add:
            peers.append(nodes[idx].ma)
        d['peers'] = wrap_list_to_json(peers)

        conf_string = template.substitute(d)
        conf_file = open(confFileName, "w")
        conf_file.write(conf_string)
        conf_file.close()
        i += 1


def set_toadd(i, n):
    toAdd = set()
    for idx in n.machine.nodes:
        if idx != i:
            toAdd.add(idx)
    if not n.hidden:
        for idx in conf_type.pubNodes:
            if idx != i:
                toAdd.add(idx)
    return toAdd


if __name__ == '__main__':
    cwd = os.getcwd()

    argc = len(sys.argv)
    if argc == 2:
        startingPort = int(sys.argv[1])
    else:
        print("Usage: %s <startingport>" % sys.argv[0])
        sys.exit(1)

    conf_type.setup_nodes(conf.nodes, startingPort)

    tmplfile = open("argconf_template.toml", 'r')
    tmplString = tmplfile.read()
    tmplfile.close()
    template = string.Template(tmplString)

    generate_confs(conf.nodes, template)
    print("Finished!")
