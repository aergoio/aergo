/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"net"
	"reflect"
	"testing"
)

func TestResolveHostDomain(t *testing.T) {
	type args struct {
		domainname string
	}
	tests := []struct {
		name    string
		args    args
		exist   bool
		wantErr bool
	}{
		{"TSucc",args{"www.google.com"},true,false},
		{"TNowhere",args{"not.in.my.aergo.io"},false,true},
		{"TWrongName",args{"!#@doigjw"},false,true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveHostDomain(tt.args.domainname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveHostDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(got) > 0 ) != tt.exist {
				t.Errorf("ResolveHostDomain() = %v, want %v", got, tt.exist)
			}
		})
	}
}

func TestResolveHostDomainLocal(t *testing.T) {
	t.Skip("skip env dependent test")
	type args struct {
		domainname string
	}
	tests := []struct {
		name    string
		args    args
		want    []net.IP
		wantErr bool
	}{
		{"TPrivate",args{"devuntu31"},[]net.IP{net.ParseIP("192.168.0.215")},false},
		{"TPrivate",args{"devuntu31.blocko.io"},[]net.IP{net.ParseIP("192.168.0.215")},false},
		{"TPrivate",args{"devuntu31ss"},nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveHostDomain(tt.args.domainname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveHostDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveHostDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name string

		in string

		wantErr  bool
		wantHost string
		wantPort string
	}{
		{"TIP4","211.34.56.78",false,"211.34.56.78",""},
		{"TIP6","fe80::dcbf:beff:fe87:e30a",false,"fe80::dcbf:beff:fe87:e30a",""},
		{"TIP6_2","::ffff:192.0.1.2",false,"::ffff:192.0.1.2",""},
		{"TFQDN","iparkmac.aergo.io",false,"iparkmac.aergo.io",""},
		{"TIP4withPort","211.34.56.78:1234",true,"211.34.56.78","1234"},
		{"TIP6withPort","[fe80::dcbf:beff:fe87:e30a]:1234",true,"fe80::dcbf:beff:fe87:e30a","1234"},
		{"TFQDNwithPort","iparkmac.aergo.io:1234",true,"iparkmac.aergo.io","1234"},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CheckAdddress(test.in)
			if (err != nil) != test.wantErr {
				t.Errorf("CheckAdddress() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr {
				if got != test.wantHost {
					t.Errorf("CheckAdddress() = host %v, wantHost %v", got, test.wantHost)
				}
			}
		})
	}
}


func TestCheckAddressType(t *testing.T) {
	tests := []struct {
		name string

		in string

		want  AddressType
	}{
		{"TIP4","211.34.56.78",AddressTypeIP},
		{"TIP6","fe80::dcbf:beff:fe87:e30a",AddressTypeIP},
		{"TIP6_2","::ffff:192.0.1.2",AddressTypeIP},
		{"TFQDN","iparkmac.aergo.io",AddressTypeFQDN},
		{"TFQDN_2","3com.com",AddressTypeFQDN},
		{"TWrongDN","3com!.com",AddressTypeError},
		{"TIP4withPort","211.34.56.78:1234",AddressTypeError},
		{"TIP6withPort","[fe80::dcbf:beff:fe87:e30a]:1234",AddressTypeError},
		{"TFQDNwithPort","iparkmac.aergo.io:1234",AddressTypeError},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckAdddressType(test.in)
				if got != test.want {
					t.Errorf("CheckAdddressType() = type %v, wantType %v", got, test.want)
				}
		})
	}
}

