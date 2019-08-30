/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// This file describe the command to generate mock objects of imported interfaces

// in aergo but outside of p2p
//go:generate sh -c "mockgen github.com/aergoio/aergo/types ChainAccessor | sed -e 's/[Pp]ackage mock_types/package p2pmock/g' > ../p2pmock/mock_chainaccessor.go"

// in aergoio
//go:generate sh -c "mockgen github.com/aergoio/aergo-actor/actor Context | sed -e 's/[Pp]ackage mock_actor/package p2pmock/g' > ../p2pmock/mock_actorcontext.go"

//go:generate sh -c "mockgen github.com/aergoio/aergo/consensus ConsensusAccessor,AergoRaftAccessor | sed -e 's/^package mock_consensus/package p2pmock/g' > ../p2pmock/mock_consensus.go"

// golang base
//go:generate sh -c "mockgen io Reader,ReadCloser,Writer,WriteCloser,ReadWriteCloser | sed -e 's/^package mock_io/package p2pmock/g'  > ../p2pmock/mock_io.go"

// libp2p
//go:generate sh -c "mockgen github.com/libp2p/go-libp2p-core Host | sed -e 's/^package mock_go_libp2p_core/package p2pmock/g' > ../p2pmock/mock_host.go"

//go:generate sh -c "mockgen github.com/libp2p/go-libp2p-core/network Stream | sed -e 's/^package mock_network/package p2pmock/g' > ../p2pmock/mock_stream.go"
