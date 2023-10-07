/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"errors"

	"github.com/aergoio/aergo/v2/types"
)

const DefaultBlockBroadcastBuffer = 5

// ListBlockStream manages server stream of listBlock RPC
type ListBlockStream struct {
	id         uint32
	awayChan   chan interface{}
	finishSend chan interface{}
	sender     chan *types.Block
	stream     types.AergoRPCService_ListBlockStreamServer
}

func NewListBlockStream(id uint32, stream types.AergoRPCService_ListBlockStreamServer) *ListBlockStream {
	return &ListBlockStream{
		id:         id,
		awayChan:   make(chan interface{}),
		sender:     make(chan *types.Block, DefaultBlockBroadcastBuffer),
		finishSend: make(chan interface{}, 1),
		stream:     stream,
	}
}

func (s *ListBlockStream) StartSend() {
	var err error
SEND_LOOP:
	for {
		select {
		case <-s.finishSend: // for priority reason
			logger.Debug().Uint32("id", s.id).Msg("finishing send ListBlock")
			break SEND_LOOP
		default:
			select {
			case <-s.finishSend:
				logger.Debug().Uint32("id", s.id).Msg("finishing send ListBlock")
				break SEND_LOOP
			case block := <-s.sender:
				err = s.stream.Send(block)
				if err != nil {
					logger.Warn().Uint32("streamId", s.id).Err(err).Msg("failed to broadcast block stream")
				} else {
					logger.Trace().Uint32("streamId", s.id).Uint64("number", block.BlockNo()).Msg("broadcast new block")
				}
			}
		}
	}
}

func (s *ListBlockStream) Send(block *types.Block) error {

	select {
	case s.sender <- block:
		return nil
	default:
		s.GoAway()
		return errors.New("stream is in stuck")
	}
}

func (s *ListBlockStream) GoAway() {
	select {
	case s.awayChan <- 0:
		// it's ok
	default:
		logger.Warn().Uint32("streamId", s.id).Msg("goAway called twice. it looks like a bug")
	}
}

// ListBlockMetaStream manages server stream of listBlock RPC
// FIXME this class has a lot of duplication to ListBlockStream
type ListBlockMetaStream struct {
	id         uint32
	awayChan   chan interface{}
	finishSend chan interface{}
	sender     chan *types.BlockMetadata
	stream     types.AergoRPCService_ListBlockMetadataStreamServer
}

func NewListBlockMetaStream(id uint32, stream types.AergoRPCService_ListBlockMetadataStreamServer) *ListBlockMetaStream {
	return &ListBlockMetaStream{
		id:         id,
		awayChan:   make(chan interface{}),
		sender:     make(chan *types.BlockMetadata, DefaultBlockBroadcastBuffer),
		finishSend: make(chan interface{}, 1),
		stream:     stream,
	}
}

func (s *ListBlockMetaStream) StartSend() {
	var err error
SEND_LOOP:
	for {
		select {
		case <-s.finishSend: // for priority reason
			logger.Debug().Uint32("id", s.id).Msg("finishing send ListBlockMetadata")
			break SEND_LOOP
		default:
			select {
			case <-s.finishSend:
				logger.Debug().Uint32("id", s.id).Msg("finishing send ListBlockMetadata")
				break SEND_LOOP
			case block := <-s.sender:
				err = s.stream.Send(block)
				if err != nil {
					logger.Warn().Uint32("streamId", s.id).Err(err).Msg("failed to broadcast block metadata stream")
				} else {
					logger.Trace().Uint32("streamId", s.id).Stringer("hash", types.LogBase58(block.Hash)).Msg("broadcast new block metadata")
				}
			}
		}
	}
}

func (s *ListBlockMetaStream) Send(block *types.BlockMetadata) error {

	select {
	case s.sender <- block:
		return nil
	default:
		s.GoAway()
		return errors.New("stream is in stuck")
	}
}

func (s *ListBlockMetaStream) GoAway() {
	select {
	case s.awayChan <- 0:
		// it's ok
	default:
		logger.Warn().Uint32("streamId", s.id).Msg("goAway called twice. it looks like a bug")
	}
}
