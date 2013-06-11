package fs

/*
#cgo LDFLAGS: -lc
#include <sys/param.h>
#include <sys/mount.h>
#include <sys/uio.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

var mountLock sync.RWMutex

func nmount(options map[string][]byte, flags int) (*MountInfo, error) {
	mountLock.Lock()
	defer mountLock.Unlock()

	toBytes, ok := options["fspath"]
	if !ok {
		return nil, fmt.Errorf("fspath non-optional")
	}

	to := string(toBytes)

	var iovs []C.struct_iovec
	var iov C.struct_iovec

	for k, value := range options {
		key := []byte(k)
		key = append(key, 0)
		value = append(value, 0)

		iov.iov_base = unsafe.Pointer(&key[0])
		iov.iov_len = C.size_t(len(key))
		iovs = append(iovs, iov)

		iov.iov_base = unsafe.Pointer(&value[0])
		iov.iov_len = C.size_t(len(value))
		iovs = append(iovs, iov)
	}

	if _, er := C.nmount(&iovs[0], C.uint(len(iovs)), C.int(flags)); er != nil {
		return nil, er
	}

	return MountInfoForPath(to)
}

// MountNullfs mounts a path on top an arbitrary mountpoint. The filesystem
// at the mountpoint is not accessible under the nullfs mount (as opposed
// to a unionfs mount). The mounted and mountee must be distinct paths or the
// kernel will complain.
//
// Requires the nullfs kernel module.
func MountNullfs(from, to string, flags int) (*MountInfo, error) {
	args := map[string][]byte{
		"fstype": []byte("nullfs"),
		"fspath": []byte(to),
		"target": []byte(from),
	}

	return nmount(args, flags)
}

// MountUnionfs mounts a path on top of an arbitrary mountpoint. The filesystem
// at the mountpoint is accessible under the unionfs mount (as opposed to
// a nullfs mount). from is considered the top layer, and all changes are made
// to it. The mounted and mountee must be distinct paths, or an error is returned.
// Avoid removing files from a unionfs, as a shadow entry will be
// made on the top filesystem and accessing the original file will be very
// difficult.
func MountUnionfs(from, to string, flags int) (*MountInfo, error) {
	args := map[string][]byte{
		"fstype": []byte("unionfs"),
		"fspath": []byte(to),
		"from":   []byte(from),
	}

	return nmount(args, flags)
}

// MountUfs mounts a normal UFS filesystem. from should be a character device
// containing a valid, *trusted* UFS filesystem.
func MountUfs(from, to string, flags int) (*MountInfo, error) {
	args := map[string][]byte{
		"fstype": []byte("ufs"),
		"fspath": []byte(to),
		"from":   []byte(from),
	}

	return nmount(args, flags)
}
