/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-addr-util"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	"github.com/stretchr/testify/assert"
)

const SampleAddrString = "/ip4/172.21.11.12/tcp/3456"
const NetAddrString = "172.21.11.12:3456"

func TestGetIP(t *testing.T) {
	addrInput, _ := ma.NewMultiaddr(SampleAddrString)

	netAddr, err := mnet.ToNetAddr(addrInput)
	if err != nil {
		t.Errorf("Invalid func %s", err.Error())
	}
	fmt.Printf("net Addr : %s", netAddr.String())
	if NetAddrString != netAddr.String() {
		t.Errorf("Expected %s, but actually %s", NetAddrString, netAddr)
	}

	addrInput, _ = ma.NewMultiaddr(SampleAddrString + "/ipfs/16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	netAddr, err = mnet.ToNetAddr(addrInput)
	if nil == err {
		t.Errorf("Error expected, but not")
	}

}
func TestLookupAddress(t *testing.T) {
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ip)

}

func TestAddrUtil(t *testing.T) {
	addrs, err := addrutil.InterfaceAddresses()
	if err != nil {
		t.Errorf("Test Error: %s", err.Error())
	}
	fmt.Printf("Addrs : %s\n", addrs)
	fmt.Println("Over")
}

func Test_debugLogReceiveMsg(t *testing.T) {
	logger := log.NewLogger("test.p2p")
	peerID, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	peer := &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: peerID},name: p2putil.ShortForm(peerID)+"@1"}
	msgID := uuid.Must(uuid.NewV4()).String()
	dummyArray := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	type args struct {
		protocol   p2pcommon.SubProtocol
		additional interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{PingRequest, nil}},
		{"int", args{PingResponse, len(msgID)}},
		{"pointer", args{StatusRequest, &logger}},
		{"array", args{StatusRequest, dummyArray}},
		{"string", args{StatusRequest, "string addition"}},
		{"obj", args{PingRequest, P2P{}}},
		{"lazy", args{PingRequest, log.DoLazyEval(func() string {
			return "Length is " + strconv.Itoa(len(dummyArray))
		})}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debugLogReceiveMsg(logger, tt.args.protocol, msgID, peer, tt.args.additional)
		})
	}
}

func Test_Encode(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		out  string
	}{
		{"TEmpty", []byte(""), ""},
		{"TNil", nil, ""},
		{"TOnlyone", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "11111111111111111111111111111111"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := enc.ToString(test.in)
			assert.Equal(t, test.out, got)
			if len(test.out) > 0 {
				gotBytes, err := enc.ToBytes(test.out)
				assert.Nil(t, err)
				assert.Equal(t, test.in, gotBytes)
			}
		})
	}
}

// experiment, and manually checked it
func TestExperiment_TimerBehavior(t *testing.T) {
	t.Skip(" experiment, do manual check")
	counter := int32(0)
	ticked := make(chan int32)
	duration := time.Millisecond * 50
	target := time.AfterFunc(duration, func() {
		fmt.Printf("Timer fired! %d \n", atomic.LoadInt32(&counter))
		atomic.AddInt32(&counter, 1)
		ticked <- counter
	})
	assert.Equal(t, int32(1), <-ticked)
	fmt.Println("Waking up")
	stopResult := target.Stop()
	assert.Equal(t, false, stopResult)
	for i := 2; i < 10; i++ {
		target.Reset(duration * 2)
		assert.Equal(t, int32(i), <-ticked)
	}
}

func TestFailNormalMapInConcurrent(t *testing.T) {
	t.Skip(" experiment, do manual check")
	iterSize := 10000
	target := make(map[string]string)
	wg := sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			target[strconv.Itoa(i)] = "var " + strconv.Itoa(i)
			if i == (iterSize >> 2) {
				wg.Done()
			}
		}
	}()

	go func() {
		wg.Wait()
		// for key, val := range target {
		// 	fmt.Printf("%s is %s\n", key, val)
		// }
		fmt.Printf("%d values after ", len(target))
		waitChan <- 0
	}()

	<-waitChan
	fmt.Printf("Remains %d\n", len(target))
}

func TestSyncMap(t *testing.T) {
	iterSize := 1000
	target := sync.Map{}
	wg := sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			target.Store(strconv.Itoa(i), "var "+strconv.Itoa(i))
			if i == (iterSize >> 2) {
				wg.Done()
			}
		}
	}()

	go func() {
		wg.Wait()
		// target.Range(func(key interface{}, val interface{}) bool {
		// 	fmt.Printf("%s is %s\n", key, val)
		// 	return true
		// })
		waitChan <- 0
	}()

	<-waitChan
	fmt.Printf("Remains %d\n", func(m *sync.Map) int {
		keys := make([]string, 0, iterSize*2)
		m.Range(func(key interface{}, val interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
		//		fmt.Println(keys)
		return len(keys)
	}(&target))
}

func TestComparePeerID(t *testing.T) {
	samplePeerID := dummyPeerID
	samplePeerID2 := dummyPeerID2
	shorterPeerID := peer.ID(string(([]byte(samplePeerID))[:len(string(samplePeerID))-1]))
	fmt.Println("Sample1", []byte(string(samplePeerID)))
	fmt.Println("Sample2", []byte(string(samplePeerID2)))
	fmt.Println("Shorter", []byte(string(shorterPeerID)))
	tests := []struct {
		name string
		p1   peer.ID
		p2   peer.ID
		want int
	}{
		{"TP1", samplePeerID, samplePeerID2, 1},
		{"TP2", samplePeerID, shorterPeerID, 1},
		{"TZ", samplePeerID, samplePeerID, 0},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComparePeerID(tt.p1, tt.p2); sign(got) != tt.want {
				t.Errorf("ComparePeerID() = %v, want %v", got, tt.want)
			}
			if got := ComparePeerID(tt.p2, tt.p1); sign(got) != tt.want*(-1) {
				t.Errorf("ComparePeerID() = %v, want %v", got, tt.want*(-1))
			}
		})
	}
}

func sign(value int) int {
	if value > 0 {
		return 1
	} else if value == 0 {
		return 0
	} else {
		return -1
	}
}

func TestLRUCache(t *testing.T) {
	size := 100
	target, _ := lru.New(10)

	var testSlice [16]byte

	t.Run("TArray", func(t *testing.T) {
		// var samples = make([]([hashSize]byte), size)
		for i := 0; i < size; i++ {
			copy(testSlice[:], uuid.Must(uuid.NewV4()).Bytes())
			target.Add(testSlice, i)

			assert.True(t, target.Len() <= 10)
		}
	})
}

