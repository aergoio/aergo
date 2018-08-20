/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/aergoio/aergo-lib/log"
	addrutil "github.com/libp2p/go-addr-util"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
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
		protocol   protocol.ID
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
			warnLogUnknownPeer(logger, tt.args.protocol, peerID)
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
