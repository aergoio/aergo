import subprocess
import sys
from enum import Enum


class Machine:
    def __init__(self, os, ip, conn, id, dirs):
        self.os = os
        self.ip = ip
        self.connInfo = conn
        if len(conn) > 0:
            self.connAddr, port = conn.rsplit(":", 1)
            self.connPort = int(port)
        self.userID = id
        self.baseDir = dirs
        self.nodes = []

    def __eq__(self, other):
        if self.os != other.os:
            return False
        if self.ip != other.ip:
            return False
        if self.connInfo != other.connInfo:
            return False
        if self.userID != other.userID:
            return False
        if self.baseDir != other.baseDir:
            return False
        return True

    def __str__(self):
        return "%s;%s;%s;%s@%s" % (self.os, self.ip, self.baseDir, self.userID, self.connInfo)


class Role(Enum):
    producer = 1
    watcher = 2
    agent = 3


class Node:
    def __init__(self, machine, role, hidden, clients=[]):
        self.machine = machine
        self.role = role
        self.hidden = hidden
        self.clients = clients


# don't touch below. These are filled after conf_type.setup_nodes() is called
bpNodes = []
agentNodes = []
pubNodes = []
peerIDs = []


def setup_nodes(nodes, starting):
    i = 0
    for n in nodes:
        n.rpcPort = starting
        n.p2pPort = starting + 1000
        n.profPort = starting + 2000
        servId = "s%02d" % (i % 100)
        idProc = subprocess.run(["cat", "keys/%s.id" % servId], capture_output=True)
        idProc.check_returncode()
        n.peerid = idProc.stdout.decode("utf-8")
        n.agent = ''
        n.producers = []
        n.ma = "/ip4/%s/tcp/%d/p2p/%s" % (n.machine.ip, n.p2pPort, n.peerid)
        n.machine.nodes.append(i)

        peerIDs.append(n.peerid)
        if n.role == Role.agent:
            agentNodes.append(i)
        elif n.role == Role.producer:
            bpNodes.append(i)
        if not n.hidden:
            pubNodes.append(i)

        starting += 1
        i += 1

    for n in nodes:

        if n.role == Role.agent:
            for c in n.clients:
                p = nodes[c]
                if p.machine != n.machine:
                    print("wrong conf.py: agent and producer is not same machine. \nagent    %s \nproducer %s " % (
                        n.machine, p.machine), file=sys.stderr)
                    sys.exit(1)
                n.producers.append(p.peerid)
                p.agent = n.peerid


def wrap_list_to_json(strs):
    sl = ['[']
    if len(strs) > 0 :
        for p in strs:
            sl.append('\n"')
            sl.append(p)
            sl.append('"')
            sl.append(', ')
        sl.pop(len(sl)-1)
    sl.append('\n]')
    return "".join(sl)

def wrap_map_to_json(dict):
    sl = ['{']
    if len(dict) > 0 :
        for key in dict:
            sl.append('\n"')
            sl.append(key)
            sl.append('": "')
            sl.append(dict[key])
            sl.append('"')
            sl.append(', ')
        sl.pop(len(sl)-1)
    sl.append('\n}')
    return "".join(sl)


def low_bool(b):
    if b:
        return "true"
    else:
        return "false"