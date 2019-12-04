package p2putil

import (
	"net"
	"testing"
)

func TestIsContainedIP(t *testing.T) {
	_, n4, err := net.ParseCIDR("192.168.1.1/24")
	if err != nil {
		t.Fatalf("wrong input: %v",err.Error())
	}
	_, n6, err := net.ParseCIDR("2001:0db8:0123:4567:89ab:cdef:1234:5678/96")
	if err != nil {
		t.Fatalf("wrong input: %v",err.Error())
	}
	nets := []*net.IPNet{n4,n6}
	type args struct {
		ip   string
		nets []*net.IPNet
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TMulti4In", args{"192.168.1.99",nets},true},
		{"TMulti6In", args{"2001:0db8:0123:4567:89ab:cdef:ab00:ac00",nets},true},
		{"TSingle4In", args{"192.168.1.99",nets[:1]},true},
		{"TSingle6In", args{"2001:0db8:0123:4567:89ab:cdef:ab00:ac00",nets[1:2]},true},
		{"TMulti4Out", args{"192.168.2.1",nets},false},
		{"TMulti6Out", args{"2001:0db8:0123:4567:89ab:cde0:ab00:ac00",nets},false},
		{"TSingle4Out", args{"192.168.2.1",nets[:1]},false},
		{"TSingle6Out", args{"2001:0db8:0123:4567:89ab:cde0:ab00:ac00",nets[1:2]},false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.args.ip)
			if got := IsContainedIP(ip, tt.args.nets); got != tt.want {
				t.Errorf("IsContainedIP() = %v, want %v", got, tt.want)
			}
		})
	}
}