/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"net"
	"regexp"
)

type AddressType int
// AddressType
const (
	AddressTypeError AddressType = iota
	AddressTypeIP
	AddressTypeFQDN
)


var privateIPBlocks []*net.IPNet
func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 link-local
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

// IGetSingleIPAddress find and get ip address of given address string. It return first ip if DNS or /etc/hosts has multiple IPs
func GetSingleIPAddress(addrStr string) (net.IP, error) {
	switch CheckAdddressType(addrStr) {
	case AddressTypeFQDN:
		ips, err := ResolveHostDomain(addrStr)
		if err != nil {
			return nil, err
		}
		return ips[0], nil
	case AddressTypeIP:
		return net.ParseIP(addrStr), nil
	default:
		return nil, InvalidAddresss
	}
}

func ResolveHostDomain(domainName string) ([]net.IP, error) {
	addrs, err := net.LookupHost(domainName)
	if err != nil || len(addrs) == 0 {
		return nil, fmt.Errorf("Could not get IPs: %v\n", err)
	}
	ips := make([]net.IP,len(addrs))
	for i, addr := range addrs {
		ips[i] = net.ParseIP(addr)
	}
	return ips, nil
}

const (
	DN = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`
	IPv4 = `^((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`
	IPv6 = `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
)
var (
	DNPattern = regexp.MustCompile(DN)

	InvalidAddresss = fmt.Errorf("invalid address")
)

func CheckAdddressType(urlStr string) AddressType {
	if ip := net.ParseIP(urlStr) ; ip != nil {
		return AddressTypeIP
	} else if DNPattern.MatchString(urlStr) {
		return AddressTypeFQDN
	} else {
		return AddressTypeError
	}
}

func CheckAdddress(urlStr string) (string, error) {
	if ip := net.ParseIP(urlStr) ; ip != nil {
		return urlStr, nil
	} else if DNPattern.MatchString(urlStr) {
		return urlStr, nil
	} else {
		return "", InvalidAddresss
	}
}

func IsExternalAddr(addrStr string) bool {
	switch CheckAdddressType(addrStr) {
	case AddressTypeIP:
		parced := net.ParseIP(addrStr)
		return !isPrivateIP(parced)
	case AddressTypeFQDN:
		return true
	default:
		return false
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
