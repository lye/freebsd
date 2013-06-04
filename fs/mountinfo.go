package fs

/* 
#cgo LDFLAGS: -lc
#include <sys/param.h>
#include <sys/ucred.h>
#include <sys/mount.h>

struct statfs* offset(struct statfs *v, int i) {
	return v + i;
}
*/
import "C"
import (
	"fmt"
	"os/user"
	"sync"
)

// MountInfo is a struct returned by the statfs system call. Accessor
// functions are provided for extracting fields -- avoid pulling them
// out directly. MountInfo contains a description of a single mounted
// filesystem.
//
// A MountInfo is a snapshot of state and may not reflect the current
// state of the filesystem.
type MountInfo C.struct_statfs

var mountInfoLock sync.RWMutex

// GetMountInfo wraps getmntinfo.
//
// GetMountInfo returns the MountInfo for all currently-mounted filesystems.
// As far as I can make out, it will never return an error save for a
// hardware fault.
func GetMountInfo() []MountInfo {
	/* getmntinfo uses an internal buffer (which cannot be freed) to store
	 * everything -- it is not threadsafe -- and we have to protect access
	 * to it ourselves. */
	mountInfoLock.Lock()
	defer mountInfoLock.Unlock()

	var tmp *C.struct_statfs
	i := int(C.getmntinfo(&tmp, 0))

	info := make([]MountInfo, i)

	for j, _ := range info {
		s := C.offset(tmp, C.int(j))
		info[j] = MountInfo(*s)
	}

	return info
}

// MountInfoForPath wraps statfs.
//
// MountInfoForPath returns information about the filesystem mounted at
// the specified path.
func MountInfoForPath(path string) (*MountInfo, error) {
	mountInfoLock.Lock()
	defer mountInfoLock.Unlock()

	var tmp C.struct_statfs

	_, er := C.statfs(C.CString(path), &tmp)
	if er != nil {
		return nil, er
	}

	info := MountInfo(tmp)

	return &info, nil
}

func (mi *MountInfo) is(lhs *MountInfo) bool {
	fsid1 := mi.FilesystemId()
	fsid2 := lhs.FilesystemId()

	return fsid1[0] == fsid2[0] && fsid1[1] == fsid2[1]
}

// Version returns f_version, the structure version number. I have no idea
// what this means.
func (mi *MountInfo) Version() uint32 {
	return uint32(mi.f_version)
}

// Type returns f_type, the type id of the filesystem.
//
// XXX: Are these stable? Or opaque?
func (mi *MountInfo) Type() uint32 {
	return uint32(mi.f_type)
}

// Flags returns f_flags, the flags the filesystem is mounted with.
func (mi *MountInfo) Flags() uint64 {
	return uint64(mi.f_flags)
}

// BlockSize returns f_bsize, the filesystem fragment size.
func (mi *MountInfo) BlockSize() uint64 {
	return uint64(mi.f_bsize)
}

// IoSize returns f_iosize, the optimal transfer block size.
func (mi *MountInfo) IoSize() uint64 {
	return uint64(mi.f_iosize)
}

// NumBlocks returns f_blocks, the total blocks in the filesystem.
func (mi *MountInfo) NumBlocks() uint64 {
	return uint64(mi.f_blocks)
}

// NumFreeBlocks returns f_bfree, the number of free blocks in the filesystem.
// This includes blocks reserved for superuser-only use (the last 10%, by default,
// I believe).
func (mi *MountInfo) NumFreeBlocks() uint64 {
	return uint64(mi.f_bfree)
}

// NumAvailBlocks returns f_bavail, the number of blocks available for use by
// normal users. This may be negative, which represents a filesystem past it's
// quota (and that only the superuser may access).
func (mi *MountInfo) NumAvailBlocks() int64 {
	return int64(mi.f_bavail)
}

// NumFileNodes returns f_files, the total number of file nodes in the 
// filesystem. This is not the number of files in the system, but effectively
// the number of inodes (both used and free).
func (mi *MountInfo) NumFileNodes() uint64 {
	return uint64(mi.f_files)
}

// NumFreeFileNodes returns f_ffree, the number of free file nodes available
// to non-superusers. This may be less than 0, which indicates non-root users
// may not create new files.
func (mi *MountInfo) NumFreeFileNodes() int64 {
	return int64(mi.f_ffree)
}

// NumSyncWrites returns f_syncwrites, the number of sync writes since the fs
// was mounted.
func (mi *MountInfo) NumSyncWrites() uint64 {
	return uint64(mi.f_syncwrites)
}

// NumAsyncWrites returns f_asyncwrites, the number of async writes since the
// fs was mounted.
func (mi *MountInfo) NumAsyncWrites() uint64 {
	return uint64(mi.f_asyncwrites)
}

// NumSyncReads returns f_syncreads, the number of sync reads since the fs
// was mounted.
func (mi *MountInfo) NumSyncReads() uint64 {
	return uint64(mi.f_syncreads)
}

// NumAsyncReads returns f_asyncreads, the number of async reads since the fs
// was mounted.
func (mi *MountInfo) NumAsyncReads() uint64 {
	return uint64(mi.f_asyncreads)
}

// OwnerId returns f_owner, the string-encoded uid of the user that mounted
// the filesystem.
func (mi *MountInfo) OwnerId() string {
	return fmt.Sprintf("%d", uint64(mi.f_owner))
}

// Owner returns the *User that mounted the filesystem.
func (mi *MountInfo) Owner() (*user.User, error) {
	return user.Lookup(mi.OwnerId())
}

// FilesystemId returns f_fsid; I have no idea what this is.
func (mi *MountInfo) FilesystemId() []int32 {
	return []int32{int32(mi.f_fsid.val[0]), int32(mi.f_fsid.val[1])}
}

// FsTypeName return f_fstypename which is a human-readable string indicating
// what filesystem type the mount is.
func (mi *MountInfo) FsTypeName() string {
	return C.GoString(&mi.f_fstypename[0])
}

// MntFromName returns f_mntfromname, the filesystem source.
func (mi *MountInfo) MntFromName() string {
	return C.GoString(&mi.f_mntfromname[0])
}

// MntToName returns f_mntonname, where the filesystem is mounted.
func (mi *MountInfo) MntToName() string {
	return C.GoString(&mi.f_mntonname[0])
}

// Unmount unmounts the filesystem.
func (mi *MountInfo) Unmount() error {
	_, er := C.unmount(&mi.f_mntonname[0], 0)
	return er
}

// IsMounted returns true iff the filesystem is currently mounted.
func (mi *MountInfo) IsMounted() bool {
	currentInfo := GetMountInfo()

	for i := range currentInfo {
		if mi.is(&currentInfo[i]) {
			return true
		}
	}

	return false
}

// TopMount returns the filesystem mounted on top of this filesystem,
// if one exists. It may return nil without error if !IsTopMount().
func (mi *MountInfo) TopMount() (*MountInfo, error) {
	top, er := MountInfoForPath(mi.MntToName())
	if er != nil {
		return nil, er
	}

	if mi.is(top) {
		return nil, nil
	}

	return top, nil
}

// IsTopMount returns true iff no other filesystem is mounted over this filesystem.
// It will additionally return false if a unionfs is mounted over this.
func (mi *MountInfo) IsTopMount() (bool, error) {
	top, er := mi.TopMount()
	if er != nil {
		return false, er
	}

	return top == nil, nil
}
