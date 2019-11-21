import conftype
from conftype import Role

testname = "p2ptest"

machines = (conftype.Machine("linux", "192.168.1.12", "192.168.1.12:22", "sg31park", "/var/blockchain/pytest"),
            conftype.Machine("mac", "192.168.1.13", "", "", "/Users/sg31park/Developer/blockchain/pytest"),
            conftype.Machine("mac", "192.168.50.110", "localhost:19999", "sg31park",
                             "/Users/sg31park/Developer/blockchain/pytest"),
            )

# total bp and agent count and organization is hardcode yet
nodes = (conftype.Node(machines[0], Role.producer, True),
         conftype.Node(machines[0], Role.producer, True),
         conftype.Node(machines[0], Role.producer, True),
         conftype.Node(machines[0], Role.agent, False, [0, 1, 2]),
         conftype.Node(machines[1], Role.producer, True),
         conftype.Node(machines[1], Role.producer, True),
         conftype.Node(machines[1], Role.agent, False, [4, 5]),
         conftype.Node(machines[1], Role.watcher, True),
         conftype.Node(machines[2], Role.producer, False),
         conftype.Node(machines[2], Role.watcher, False),
         )

polarises = ()
#polarises = ("/ip4/192.168.1.12/tcp/8916/p2p/16Uiu2HAmJCmxe7CrgTbJBgzyG8rx5Z5vybXPWQHHGQ7aRJfBsoFs", "/ip4/192.168.1.13/tcp/8916/p2p/16Uiu2HAm1avWH56QiCXiVNKVerPYZNoqYKV3FxwbAtHU22iJKXZ3")

# genesis info
gen_magic = "test.hayarobi.aergo"
holders = (0, 4, 8)
