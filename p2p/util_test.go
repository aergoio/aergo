package p2p

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru"

	"github.com/aergoio/aergo-lib/log"
	addrutil "github.com/libp2p/go-addr-util"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	uuid "github.com/satori/go.uuid"
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
	msgID := uuid.Must(uuid.NewV4()).String()
	dummyArray := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	type args struct {
		protocol   SubProtocol
		additional interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{pingRequest, nil}},
		{"int", args{pingResponse, len(msgID)}},
		{"pointer", args{statusRequest, &logger}},
		{"array", args{getMissingRequest, dummyArray}},
		{"string", args{getMissingResponse, "string addition"}},
		{"obj", args{pingRequest, P2P{}}},
		{"lazy", args{pingRequest, log.DoLazyEval(func() string {
			return "Length is " + strconv.Itoa(len(dummyArray))
		})}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debugLogReceiveMsg(logger, tt.args.protocol, msgID, peerID, tt.args.additional)
		})
	}
}

func TestTimerBehavior(t *testing.T) {
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

func FailTestNormalMapInConcurrent(t *testing.T) {
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

type MockLogger struct {
}

func TestComparePeerID(t *testing.T) {
	samplePeerID, _ := peer.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")
	samplePeerID2, _ := peer.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
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
