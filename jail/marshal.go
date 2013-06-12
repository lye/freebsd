package jail

/*
#cgo LDFLAGS: -ljail
#include <sys/param.h>
#include <sys/jail.h>
#include <jail.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"unsafe"
)

var (
	intType, stringType, ipType, ipSliceType, boolType reflect.Type
	paramTypeMapping                                   map[string]reflect.Type
)

func jailParamType(name string) reflect.Type {
	if ty, ok := paramTypeMapping[name]; ok {
		return ty
	}

	return nil
}

func deref(iface interface{}) reflect.Value {
	val := reflect.ValueOf(iface)

	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	return val
}

func jailParamTypeCoerce(name string, iface interface{}) (ptr unsafe.Pointer, ptrLen int, er error) {
	val := deref(iface)
	kind := val.Kind()

	ty := jailParamType(name)
	if ty == nil {
		return nil, 0, fmt.Errorf("Invalid parameter `%s'", name)
	}

	if ty == intType {
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 {
			return nil, 0, fmt.Errorf("Parameter `%s' must be a string", name)
		}

		return unsafe.Pointer(val.UnsafeAddr()), int(val.Type().Size()), nil

	} else if ty == stringType {
		if kind != reflect.String {
			return nil, 0, fmt.Errorf("Parameter `%s' must be a string", name)
		}

		sval := []byte(val.String())
		return unsafe.Pointer(&sval[0]), len(sval), nil

	} else if ty == boolType {
		if kind != reflect.Bool {
			return nil, 0, fmt.Errorf("Parameter `%s' must be a bool", name)
		}

		bval := val.Bool()
		return unsafe.Pointer(&bval), int(unsafe.Sizeof(bval)), nil

	} else if ty == ipType {
		if val, ok := val.Interface().(net.IP); ok {
			sval := []byte(val)
			return unsafe.Pointer(&sval[0]), len(sval), nil
		}

		return nil, 0, fmt.Errorf("Parameter `%s' must be a net.IP", name)

	} else if ty == ipSliceType {
		buf := bytes.Buffer{}

		if ips, ok := val.Interface().([]net.IP); ok {
			for _, ip := range ips {
				buf.Write([]byte(ip))
			}

		} else {
			return nil, 0, fmt.Errorf("Parameter `%s' must be a []net.IP or []*net.IP", name)
		}

		if buf.Len() > 0 {
			return unsafe.Pointer(&buf.Bytes()[0]), buf.Len(), nil

		} else {
			/* XXX: Not sure if this is how we encode an empty slice? */
			return nil, 0, nil
		}
	}

	return nil, 0, fmt.Errorf("Unknown type for parameter `%s'", name)
}

func jailParamTypeExtract(name string, val *C.struct_jailparam, out interface{}) error {
	outVal := deref(out)
	kind := outVal.Kind()

	ty := jailParamType(name)
	if ty == nil {
		return fmt.Errorf("Unknown parameter `%s'", name)
	}

	if ty == intType {
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 {
			return fmt.Errorf("Parameter %s must be an int")
		}

		intVal := int64(*(*C.int)(val.jp_value))
		outVal.SetInt(intVal)

	} else if ty == stringType {
		if kind != reflect.String {
			return fmt.Errorf("Parameter %s must be a string")
		}

		strVal := string(C.GoBytes(val.jp_value, C.int(val.jp_valuelen)))
		strVal = strings.Trim(strVal, "\x00")
		outVal.SetString(strVal)

	} else if ty == boolType {
		if kind != reflect.Bool {
			return fmt.Errorf("Parameter %s must be a bool")
		}

		boolVal := byte(*(*C.char)(val.jp_value))
		outVal.SetBool(boolVal != 0)

	} else if ty == ipType || ty == ipSliceType {
		buf := bytes.NewBuffer(C.GoBytes(val.jp_value, C.int(val.jp_valuelen)))
		newSlice := reflect.ValueOf([]net.IP{})
		stride := 0

		if name == "ip4.addr" {
			stride = 4

		} else if name == "ip6.addr" {
			stride = 16

		} else {
			return fmt.Errorf("Parameter `%s' cannot be coerced to net.IP's", name)
		}

		for buf.Len() >= stride {
			b := make([]byte, stride)

			if _, er := buf.Read(b); er != nil {
				return er
			}

			fmt.Printf("%#v\n", b)

			ip := net.IP(b)
			newSlice = reflect.Append(newSlice, reflect.ValueOf(ip))
		}

		if _, ok := outVal.Interface().([]net.IP); ok {
			outVal.Set(newSlice)

		} else if _, ok := outVal.Interface().(net.IP); ok {
			outVal.Set(newSlice.Index(0))

		} else {
			fmt.Printf("Parameter `%s' must be a net.IP or []net.IP", name)
		}

	} else {
		return fmt.Errorf("Unknown type for parameter `%s'", name)
	}

	return nil
}
