package jail

import (
	"net"
	"reflect"
)

func init() {
	intType = reflect.TypeOf(int(1))
	stringType = reflect.TypeOf("")
	ipType = reflect.TypeOf(net.IP{})
	ipSliceType = reflect.TypeOf([]net.IP{})
	boolType = reflect.TypeOf(true)

	paramTypeMapping = map[string]reflect.Type{
		"jid":     intType,
		"lastjid": intType,
		"name":    stringType,
		"path":    stringType,

		"ip4":          stringType,
		"ip4.addr":     ipSliceType,
		"ip4.saddrsel": boolType,

		"ip6":          stringType,
		"ip6.addr":     ipSliceType,
		"ip6.saddrsel": boolType,

		"host":            stringType,
		"host.hostname":   stringType,
		"host.domainname": stringType,
		"host.hostuuid":   stringType,
		"host.hostid":     stringType,

		"securelevel": intType,

		"children.max": intType,
		"children.cur": intType,
		"parent":       intType,

		"enforce_statfs": boolType,
		"persist":        boolType,
		"cpuset.id":      intType,
		"dying":          boolType,

		"allow.set_hostname": boolType,
		"allow.sysvipc":      boolType,
		"allow.raw_sockets":  boolType,
		"allow.chflags":      boolType,
		"allow.mount":        boolType,
		"allow.quotas":       boolType,
		"allow.socket_af":    boolType,
	}
}
