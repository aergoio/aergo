/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
	lru "github.com/hashicorp/golang-lru"
	addrutil "github.com/libp2p/go-addr-util"
	ma "github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
	"github.com/stretchr/testify/assert"
)

const SampleAddrString = "/ip4/172.21.11.12/tcp/3456"
const NetAddrString = "172.21.11.12:3456"

func TestGetIP(t *testing.T) {
	t.Skip("skip test cause infinite loop in Goland IDE")

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
	ip, err := ExternalIP()
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

/*
	func Test_debugLogReceiveMsg(t *testing.T) {
		logger := log.NewLogger("test.p2p")
		peerID, _ := types.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
		peer := &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: peerID}, name: ShortForm(peerID) + "@1"}
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
			{"nil", args{subproto.PingRequest, nil}},
			{"int", args{subproto.PingResponse, len(msgID)}},
			{"pointer", args{subproto.StatusRequest, &logger}},
			{"array", args{subproto.StatusRequest, dummyArray}},
			{"string", args{subproto.StatusRequest, "string addition"}},
			{"obj", args{subproto.PingRequest, P2P{}}},
			{"lazy", args{subproto.PingRequest, log.DoLazyEval(func() string {
				return "Length is " + strconv.Itoa(len(dummyArray))
			})}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				DebugLogReceiveMsg(logger, tt.args.protocol, msgID, peer, tt.args.additional)
			})
		}
	}
*/
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

var dummyPeerID types.PeerID
var dummyPeerID2 types.PeerID

func init() {
	dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyPeerID2, _ = types.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")
}

func TestComparePeerID(t *testing.T) {
	samplePeerID := dummyPeerID
	samplePeerID2 := dummyPeerID2
	shorterPeerID := types.PeerID(string(([]byte(samplePeerID))[:len(string(samplePeerID))-1]))
	fmt.Println("Sample1", []byte(string(samplePeerID)))
	fmt.Println("Sample2", []byte(string(samplePeerID2)))
	fmt.Println("Shorter", []byte(string(shorterPeerID)))
	tests := []struct {
		name string
		p1   types.PeerID
		p2   types.PeerID
		want int
	}{
		{"TP1", samplePeerID, samplePeerID2, 1},
		{"TP2", samplePeerID, shorterPeerID, 1},
		{"TZ", samplePeerID, samplePeerID, 0},
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

func TestParseURI(t *testing.T) {
	tests := []struct {
		name string

		arg string

		wantErr  bool
		wantIP   bool
		wantPort bool
	}{
		{"Thttp", "http://172.121.33.4:1111", false, true, true},
		{"Thttps", "https://172.121.33.4:1111", false, true, true},
		{"TMultiAddr", "/ip4/172.121.33.4/tcp/1111", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := url.ParseRequestURI(tt.arg)
			if err != nil {
				t.Logf("Got addr %v", got)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequestURI() err %v , wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if (len(got.Host) > 0) != tt.wantIP {
					t.Errorf("ParseRequestURI() Host %v , want %v", got.Host, tt.wantIP)
				}
				if (len(got.Port()) > 0) != tt.wantPort {
					t.Errorf("ParseRequestURI() Port %v , want %v", got.Port(), tt.wantPort)
				}
			}

		})
	}
}

func TestExternalIP(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"T1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := ExternalIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExternalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("Got IP %v", ip.String())
		})
	}
}

func Test_getValidIP(t *testing.T) {
	loIP4 := "127.0.0.1"
	llIP4 := "169.254.0.2"
	priIP4 := "192.168.1.3"
	pubIP4 := "116.127.31.76"
	loIP6 := "::1"
	llIP6 := "fe80::"
	priIP6 := "192.168.1.3"
	//pubIP6:="116.127.31.76"
	wrapped := "::ffff:df01:1111"

	tests := []struct {
		name string
		args []string
		want net.IP
	}{
		{"TEmpty", []string{}, nil},
		{"TLo4", []string{loIP4}, nil},
		{"TLo6", []string{loIP6}, nil},
		{"TLoLL", []string{loIP4, llIP4}, nil},
		{"TLoLL6", []string{loIP6, llIP6}, nil},
		{"TLL4", []string{llIP4}, nil},
		{"TLL6", []string{llIP6}, nil},
		{"T4Only", []string{pubIP4}, net.ParseIP(pubIP4)},
		{"TLo4Uo6", []string{loIP4, priIP6}, net.ParseIP(priIP6)},
		{"TLL6P4", []string{priIP4, llIP6}, net.ParseIP(priIP4)},
		{"TP4LL6", []string{llIP6, priIP4}, net.ParseIP(priIP4)},
		{"TWrapped6", []string{wrapped}, net.ParseIP(wrapped)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrs := []net.Addr{}
			for _, ipstr := range tt.args {
				ip := net.ParseIP(ipstr)
				if ip == nil {
					t.Fatalf("invalid ip sample %v", ipstr)
				}
				addrs = append(addrs, &net.IPAddr{IP: ip, Zone: "ip+net"})
			}
			if got := getValidIP(addrs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValidIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonotonicClocks(t *testing.T) {
	mt := time.Now()
	mt2 := mt.Add(time.Hour)
	wt := mt.Truncate(0)
	wt2 := wt.Add(time.Hour)

	if !mt.Equal(wt) {
		t.Errorf("Equal(): Monotonic and wall clock differ! %v and %v ", mt, wt)
	}
	if !mt2.Equal(wt2) {
		t.Errorf("Monotonic and wall clock differ! %v and %v ", mt, wt)
	}
	if !reflect.DeepEqual(mt, wt) {
		// this is expected situation. you should compare clocks by Equal() method, not by reflection.
	}
}
