package netif

import (
	"strings"
	"testing"
)

func TestLocalAddrs(t *testing.T) {
	addrs, er := EnumAddrs()
	if er != nil {
		t.Fatal(er)
	}

	/* These aren't really checking "loopback" addresses; I just can't
	 * remember what these are actually called. */
	sawLoopback4 := false
	sawLoopback6 := false
	allIPs := []string{}

	for _, addr := range addrs {
		if addr.String() == "127.0.0.1" {
			sawLoopback4 = true
		}

		if addr.String() == "::1" {
			sawLoopback6 = true
		}

		allIPs = append(allIPs, addr.String())
	}

	if !sawLoopback4 || !sawLoopback6 {
		if !sawLoopback4 {
			t.Errorf("Did not see an ipv4 loopback address")
		}

		if !sawLoopback6 {
			t.Errorf("Did not see an ipv6 loopback address")
		}

		t.Logf("Saw addresses: %s", strings.Join(allIPs, ", "))
	}
}
