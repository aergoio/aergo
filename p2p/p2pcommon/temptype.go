/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// This file describe the command to generate mock objects of imported interfaces
//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/p2p/p2pcommon NTContainer,NetworkTransport | gsed -e 's/[Pp]ackage mock_p2pcommon/package p2pmock/g' > p2p/p2pmock/mock_networktransport.go"

// in aergo but outside of p2p
//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/types ChainAccessor | sed -e 's/[Pp]ackage mock_types/package p2pmock/g' > ../p2pmock/mock_chainaccessor.go"

//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/consensus ConsensusAccessor,AergoRaftAccessor | sed -e 's/^package mock_consensus/package p2pmock/g' > ../p2pmock/mock_consensus.go"

//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/types AergoRPCService_ListBlockStreamServer | sed -e 's/^package mock_types/package p2pmock/g' > ../p2pmock/mock_protobuf.go"

// in aergoio
//go:generate sh -c "mockgen github.com/aergoio/aergo-actor/actor Context | sed -e 's/[Pp]ackage mock_actor/package p2pmock/g' > ../p2pmock/mock_actorcontext.go"
//go:generate sh -c "mockgen github.com/aergoio/aergo-actor/component ICompSyncRequester | sed -e 's/[Pp]ackage mock_actor/package p2pmock/g' > ../p2pmock/mock_actorcontext.go"

// golang base
//go:generate sh -c "mockgen io Reader,ReadCloser,Writer,WriteCloser,ReadWriteCloser | sed -e 's/^package mock_io/package p2pmock/g'  > ../p2pmock/mock_io.go"

// libp2p
//go:generate sh -c "mockgen github.com/libp2p/go-libp2p-core Host | sed -e 's/^package mock_go_libp2p_core/package p2pmock/g' > ../p2pmock/mock_host.go"

//go:generate sh -c "mockgen github.com/libp2p/go-libp2p-core/network Stream,Conn | sed -e 's/^package mock_network/package p2pmock/g' > ../p2pmock/mock_stream.go"
