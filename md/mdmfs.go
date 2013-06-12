package md

/*
#cgo LDFLAGS: -lc
#include <sys/types.h>
#include <sys/ioctl.h>
#include <sys/mdioctl.h>
#include <sys/param.h>
#include <stdlib.h>
#include <strings.h>
#include <fcntl.h>

int hack_ioctl1(int d, unsigned long request, void *arg1) {
	return ioctl(d, request, arg1);
}

int hack_open1(const char *path, int flags, void *arg1) {
	return open(path, flags, arg1);
}

*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

var (
	ErrDeviceAlreadyAttached = errors.New("md: device already attached")
	ErrDeviceNotAttached     = errors.New("md: device not attached")
)

// MDDev wraps an md_ioctl with some additional state-tracking metadata. It
// represents a possible md device (either vnode- or malloc-backed) and can
// be Attached/Detached at whim.
type MDDev struct {
	attached bool
	path     string
	mdio     C.struct_md_ioctl
}

// NewVnodeMD creates a new vnode-backed MDDev using the file specified by
// path as the backing. The file should exist, and be ftruncate'd to the desired
// length beforehand.
func NewVnodeMD(path string) (*MDDev, error) {
	f, er := os.Open(path)
	if er != nil {
		return nil, er
	}
	defer f.Close()

	fi, er := f.Stat()
	if er != nil {
		return nil, er
	}

	md := &MDDev{}
	md.mdio.md_type = C.MD_VNODE
	md.mdio.md_options = C.MD_CLUSTER | C.MD_AUTOUNIT | C.MD_COMPRESS
	md.mdio.md_mediasize = C.off_t(fi.Size())

	md.path = path

	return md, nil
}

// NewMallocMD creates a new malloc-backed MDDev. Avoid using thisi for large
// allocations, as some kernels will behave strangely with large md reservations
// while under memory pressure.
func NewMallocMD(size int) (*MDDev, error) {
	md := new(MDDev)
	md.mdio.md_type = C.MD_MALLOC
	md.mdio.md_options = C.MD_AUTOUNIT | C.MD_COMPRESS
	md.mdio.md_mediasize = C.off_t(size)
	return md, nil
}

// NewSwapMD creates a new swap-backed MDDev. "Swap" in this context doesn't
// necessarily mean disk, rather, unused buffer memory that includes any 
// disk-based swap.
func NewSwapMD(size int) (*MDDev, error) {
	md := new(MDDev)
	md.mdio.md_type = C.MD_SWAP
	md.mdio.md_options = C.MD_AUTOUNIT | C.MD_COMPRESS
	md.mdio.md_mediasize = C.off_t(size)
	return md, nil
}

// Attach allocates the MDDev, if not already allocated.
func (md *MDDev) Attach() error {
	if md.attached {
		return ErrDeviceAlreadyAttached
	}

	mdctl := "/dev/" + C.MDCTL_NAME

	f, er := os.OpenFile(mdctl, os.O_RDWR, 0)
	if er != nil {
		return er
	}
	defer f.Close()

	fd := C.int(f.Fd())

	if md.path != "" {
		md.mdio.md_file = C.CString(md.path)
		defer C.free(unsafe.Pointer(md.mdio.md_file))
	}

	_, er = C.hack_ioctl1(fd, C.MDIOCATTACH, unsafe.Pointer(&md.mdio))
	if er != nil {
		return er
	}

	md.attached = true
	return nil
}

// Detach frees the MDDev. This cannot be called unless all geom providers using
// the device are closed (AFAIK). Any resources used by this device (swap/ram/etc)
// are freed.
//
// XXX: Is the backing vnode synchronized?
func (md *MDDev) Detach() error {
	if !md.attached {
		return ErrDeviceNotAttached
	}

	mdctl := "/dev/" + C.MDCTL_NAME

	f, er := os.OpenFile(mdctl, os.O_RDWR, 0)
	if er != nil {
		return er
	}
	defer f.Close()

	fd := C.int(f.Fd())

	if md.path != "" {
		md.mdio.md_file = C.CString(md.path)
		defer C.free(unsafe.Pointer(md.mdio.md_file))
	}

	tmpMdio := C.struct_md_ioctl{}
	tmpMdio.md_unit = md.mdio.md_unit

	_, er = C.hack_ioctl1(fd, C.MDIOCDETACH, unsafe.Pointer(&tmpMdio))
	if er != nil {
		return er
	}

	return nil
}

// Attached returns true iff the device is currently attached.
func (md *MDDev) Attached() bool {
	return md.attached
}

// Path returns the path to the backing file, iff the device is vnode-backed.
func (md *MDDev) Path() string {
	return md.path
}

// MallocBacked returns true iff the MDDev is malloc-backed.
func (md *MDDev) MallocBacked() bool {
	return md.mdio.md_type == C.MD_MALLOC
}

// VnodeBacked returns true iff the MDDev is vnode-backed (e.g., is a file).
func (md *MDDev) VnodeBacked() bool {
	return md.mdio.md_type == C.MD_VNODE
}

// DevicePath returns the path to the md device, which can then be mounted as
// a filesystem.
func (md *MDDev) DevicePath() (string, error) {
	if !md.attached {
		return "", ErrDeviceNotAttached
	}

	return fmt.Sprintf("/dev/md%d", int(md.mdio.md_unit)), nil
}
