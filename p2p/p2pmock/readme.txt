Examples to generate mock class

1. with no flags : it must be followed by manual editing to correct the package of generated mock class
mockgen github.com/aergoio/aergo/p2p/p2pcommon HSHandlerFactory  > p2p/p2pmock/mock_hsfactory.go

2. with flags : generate mocks of all interface in single file
mockgen -source=p2p/p2pcommon/pool.go -package=p2pmock -destination=p2p/p2pmock/mock_peerfinder.go

3. with flags others : can select the classes (exclude a class) in single file, setting class mapping is too tedious
mockgen -source=p2p/p2pcommon/pool.go -mock_names=WaitingPeerManager=MockWaitingPeerManager  -package=p2pmock -destination=p2p/p2pmock/mock_peerfinder.go

