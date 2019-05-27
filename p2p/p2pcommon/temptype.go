/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// This file describe the command to generate mock objects of imported interfaces

// libp2p
// mockgen github.com/libp2p/go-libp2p-core/network Stream | sed -e 's/^package mock_network/package p2pmock/g' > p2p/p2pmock/mock_stream.go
// mockgen github.com/libp2p/go-libp2p-core Host | sed -e 's/^package mock_go_libp2p_core/package p2pmock/g' > p2p/p2pmock/mock_host.go
