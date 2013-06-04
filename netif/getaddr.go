package netif

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <netinet/in.h>
*/
import "C"
import (
	"net"
	"unsafe"
)

// EnumAddrs returns an enumeration of all IP addresses on all network interfaces.
// Only IPv4 and IPv6 addresses are enumerated.
func EnumAddrs() ([]net.IP, error) {
	var addrs, addr *C.struct_ifaddrs

	if _, er := C.getifaddrs(&addrs); er != nil {
		return nil, er
	}

	ips := []net.IP{}

	for addr = addrs; addr != nil; addr = addr.ifa_next {
		if addr.ifa_addr == nil {
			continue
		}

		if addr.ifa_addr.sa_family == C.AF_INET {
			sockaddr := (*C.struct_sockaddr_in)(unsafe.Pointer(addr.ifa_addr))
			inaddr := sockaddr.sin_addr
			ip := C.GoBytes(unsafe.Pointer(&inaddr), 4)
			ips = append(ips, net.IP(ip))

		} else if addr.ifa_addr.sa_family == C.AF_INET6 {
			sockaddr := (*C.struct_sockaddr_in6)(unsafe.Pointer(addr.ifa_addr))
			inaddr := sockaddr.sin6_addr
			ip := C.GoBytes(unsafe.Pointer(&inaddr), 16)
			ips = append(ips, net.IP(ip))
		}
	}

	return ips, nil
}
