/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
package subproto

import (
	"fmt"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
)

func BenchmarkArrayKey(b *testing.B) {
	size := 100000
	const hashSize = 32
	var samples = make([]([hashSize]byte), size)
	for i := 0; i < size; i++ {
		copy(samples[i][:], uuid.Must(uuid.NewV4()).Bytes())
		copy(samples[i][16:], uuid.Must(uuid.NewV4()).Bytes())
	}

	b.Run("BArray", func(b *testing.B) {
		var keyArr [hashSize]byte
		startTime := time.Now()
		fmt.Printf("P1 in byte array\n")
		target := make(map[[hashSize]byte]int)
		for i := 0; i < size; i++ {
			copy(keyArr[:], samples[i][:])
			target[keyArr] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec \n", endTime.Sub(startTime).Seconds())
	})

	b.Run("Bbase64", func(b *testing.B) {
		startTime := time.Now()
		fmt.Printf("P2 in base64\n")
		target2 := make(map[string]int)
		for i := 0; i < size; i++ {
			target2[enc.ToString(samples[i][:])] = i
		}
		endTime := time.Now()
		fmt.Printf("Takes %f sec\n", endTime.Sub(startTime).Seconds())

	})

}

type MempoolRspTxCountMatcher struct {
	matchCnt int
}

func (tcm MempoolRspTxCountMatcher) Matches(x interface{}) bool {
	m, ok := x.(*message.MemPoolExistExRsp)
	if !ok {
		return false
	}
	return tcm.matchCnt == len(m.Txs)
}

func (tcm MempoolRspTxCountMatcher) String() string {
	return fmt.Sprintf("tx count = %d", tcm.matchCnt)
}

type TxIDCntMatcher struct {
	matchCnt int
}

func (scm TxIDCntMatcher) Matches(x interface{}) bool {
	m, ok := x.([]types.TxID)
	if !ok {
		return false
	}
	return scm.matchCnt == len(m)
}

func (scm TxIDCntMatcher) String() string {
	return fmt.Sprintf("len(slice) = %d", scm.matchCnt)
}

type WantErrMatcher struct {
	wantErr bool
}

func (tcm WantErrMatcher) Matches(x interface{}) bool {
	m, ok := x.(*error)
	if !ok {
		return false
	}
	return tcm.wantErr == (m != nil)
}

func (tcm WantErrMatcher) String() string {
	return fmt.Sprintf("want error = %v", tcm.wantErr)
}

func TestNewTxReqHandler(t *testing.T) {
	logger := log.NewLogger("p2p.test")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := p2pmock.NewMockPeerManager(ctrl)
	mockActor := p2pmock.NewMockActorService(ctrl)
	mockPeer := p2pmock.NewMockRemotePeer(ctrl)
	mockSM := p2pmock.NewMockSyncManager(ctrl)

	got := NewTxReqHandler(mockPM, mockSM, mockPeer, logger, mockActor)
	if got.pm != mockPM {
		t.Errorf("NewTxReqHandler() member %v is nil ", mockSM)
	}
	if got.actor != mockActor {
		t.Errorf("NewTxReqHandler() member %v is nil ", mockSM)
	}
	if got.sm != mockSM {
		t.Errorf("NewTxReqHandler() member %v is nil ", mockSM)
	}
	if got.peer != mockPeer {
		t.Errorf("NewTxReqHandler() member %v is nil ", mockSM)
	}
}
