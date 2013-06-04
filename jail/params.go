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
	"fmt"
	"unsafe"
)

type jailParamList struct {
	params      []C.struct_jailparam
	nameMapping map[string]int
}

func (jpps *jailParamList) bindParameter(name string, value interface{}) error {
	valuePtr, valueLen, er := jailParamTypeCoerce(name, value)
	if er != nil {
		return er
	}

	if er := jpps.bindOutput(name); er != nil {
		return er
	}

	jpp := &jpps.params[len(jpps.params)-1]
	jpp.jp_value = valuePtr
	jpp.jp_valuelen = C.size_t(valueLen)
	jpp.jp_flags |= C.JP_RAWVALUE

	return nil
}

func (jpps *jailParamList) bindParameters(params map[string]interface{}) error {
	for name, value := range params {
		if er := jpps.bindParameter(name, value); er != nil {
			return er
		}
	}

	return nil
}

func (jpps *jailParamList) bindOutput(name string) error {
	if jpps.nameMapping == nil {
		jpps.nameMapping = map[string]int{}
	}

	if _, ok := jpps.nameMapping[name]; ok {
		return fmt.Errorf("Cannot bind parameter `%s' twice", name)
	}

	jpps.nameMapping[name] = len(jpps.params)

	nameStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameStr))

	var jpp C.struct_jailparam
	if _, er := C.jailparam_init(&jpp, nameStr); er != nil {
		return er
	}

	jpps.params = append(jpps.params, jpp)
	return nil
}

func (jpps *jailParamList) bindOutputs(names ...string) error {
	for i := range names {
		if er := jpps.bindOutput(names[i]); er != nil {
			return er
		}
	}

	return nil
}

func (jpps *jailParamList) grabOutput(name string, out interface{}) error {
	idx, ok := jpps.nameMapping[name]
	if !ok {
		return fmt.Errorf("Parameter `%s' not in passed jailParamList", name)
	}

	jpp := &jpps.params[idx]

	return jailParamTypeExtract(name, jpp, out)
}

func (jpps *jailParamList) numParams() C.uint {
	return C.uint(len(jpps.params))
}

func (jpps *jailParamList) release() {
	if len(jpps.params) > 0 {
		C.jailparam_free(&jpps.params[0], C.uint(len(jpps.params)))
	}
}
