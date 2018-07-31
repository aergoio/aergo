/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package experiments

import (
	"fmt"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
)

func Test_Addr(t *testing.T) {
	// construct from a string (err signals parse failure)
	m1, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234")
	fmt.Println("M1 " + m1.String())
	// construct from bytes (err signals parse failure)
	m2, err := ma.NewMultiaddrBytes(m1.Bytes())

	// true
	if m1.String() != ("/ip4/127.0.0.1/udp/1234") {
		t.Errorf("invalide parsing ")
	}
	if m1.String() != m2.String() {
		t.Errorf("invalide parsing ")
	}
	// get the multiaddr protocol description objects
	protocols := m1.Protocols()
	fmt.Println(protocols)
	// []Protocol{
	//   Protocol{ Code: 4, Name: 'ip4', Size: 32},
	//   Protocol{ Code: 17, Name: 'udp', Size: 16},
	// }
	toDecap, err := ma.NewMultiaddr("/sctp/5678")
	if err != nil {
		t.Errorf("Wrong addr string %s , %s", "/sctp/5678", err.Error())
	}
	encResult := m1.Encapsulate(toDecap)
	fmt.Println("encap result " + encResult.String())
	// <Multiaddr /ip4/127.0.0.1/udp/1234/sctp/5678>
	udp, err := ma.NewMultiaddr("/udp/9876")
	if err != nil {
		t.Errorf("Wrong addr string %s , %s", "/udp", err.Error())
	}
	decapReult := m1.Decapsulate(udp) // up to + inc last occurrence of subaddr
	fmt.Println("decap result " + decapReult.String())

	// <Multiaddr /ip4/127.0.0.1>

}
