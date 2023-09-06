Examples to generate mock class

1. with reflection (and no flag is allowed) : It can generate mock of outside of current source tree, but has drawback that cannot set output package. it must be followed by manual editing or other way to correct the package of generated mock class
mockgen github.com/aergoio/aergo/v2/p2p/p2pcommon HSHandlerFactory  > p2p/p2pmock/mock_hsfactory.go


1-1. correct package name using GNU sed
mockgen github.com/libp2p/go-libp2p-core Host | gsed -e 's/^package mock_[a-zA-Z0-9_]\+/package p2pmock/g' > p2p/p2pmock/mock_host.go

NOTE: There is no suitable common way to catch regexp + (match one or more) on multiple OS environment. On linux gnu sed, the flag to enable extended regxp is -r, but on macos bsd sed, it is -E.

2. with flags : generate mocks of all interface in single file
mockgen -source=p2p/p2pcommon/pool.go -package=p2pmock -destination=p2p/p2pmock/mock_peerfinder.go

3. with flags others : can select the classes (exclude a class) in single file, setting class mapping is too tedious
mockgen -source=p2p/p2pcommon/pool.go -mock_names=WaitingPeerManager=MockWaitingPeerManager  -package=p2pmock -destination=p2p/p2pmock/mock_peerfinder.go


# Manually generated mock classes
The generate decriptions of these mock objects are in p2p/p2pcommon/temptypes.go . So you can use such like `go generate ./p2p/p2pcommon/temptypes.go` command.

# mock files which are not generated automatically by go generate ./p2p
mockgen github.com/aergoio/aergo/v2/consensus ConsensusAccessor,AergoRaftAccessor | gsed -e 's/^package mock_[a-zA-Z0-9_]\+/package p2pmock/g' > p2p/p2pmock/mock_consensus.go

mockgen -source=types/blockchain.go -package=p2pmock -destination=p2p/p2pmock/mock_chainaccessor.go

mockgen io Reader,ReadCloser,Writer,WriteCloser,ReadWriteCloser > p2p/p2pmock/mock_io.go | gsed -e 's/^package mock_[a-zA-Z0-9_]\+/package p2pmock/g'  > p2p/p2pmock/mock_io.go

mockgen github.com/aergoio/aergo/v2/types ChainAccessor | sed -e 's/^package mock_mock_[a-zA-Z0-9_]\+/package p2pmock/g' > ../p2pmock/mock_chainaccessor.go