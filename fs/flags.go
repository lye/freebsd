package fs

/*
#include <sys/param.h>
#include <sys/mount.h>
*/
import "C"

const (
	// The file system should be treated as read-only; even the
	// super-user may not write on it.  Specifying MNT_UPDATE
	// without this option will upgrade a read-only file system
	// to read/write.
	MntRdOnly = C.MNT_RDONLY

	// Do not allow files to be executed from the file system.
	MntNoExec = C.MNT_NOEXEC

	// Do not honor setuid or setgid bits on files when executing them.  
	// This flag is set automatically when the caller is not the super-user.
	MntNoSuid = C.MNT_NOSUID

	// Disable update of file access times.
	MntNoATime = C.MNT_NOATIME

	// Create a snapshot of the file system.  This is currently
	// only supported on UFS2 file systems, see mksnap_ffs(8)
	// for more information.
	MntSnapshot = C.MNT_SNAPSHOT

	// Directories with the SUID bit set chown new files to
	// their own owner.  This flag requires the SUIDDIR option
	// to have been compiled into the kernel to have any
	// effect.  See the mount(8) and chmod(2) pages for more
	// information.
	MntSuidDir = C.MNT_SUIDDIR

	// All I/O to the file system should be done synchronously.
	MntSynchronous = C.MNT_SYNCHRONOUS

	// All I/O to the file system should be done asynchronously.
	MntAsync = C.MNT_ASYNC

	// Force a read-write mount even if the file system appears
	// to be unclean.  Dangerous.  Together with MNT_UPDATE and
	// MNT_RDONLY, specify that the file system is to be
	// forcibly downgraded to a read-only mount even if some
	// files are open for writing.
	MntForce = C.MNT_FORCE

	// Disable read clustering.
	MntNoClusterR = C.MNT_NOCLUSTERR

	// Disable write clustering.
	MntNoClusterW = C.MNT_NOCLUSTERW

	// The flag MNT_UPDATE indicates that the mount command is being applied to
	// an already mounted file system.  This allows the mount flags to be
	// changed without requiring that the file system be unmounted and
	// remounted.  Some file systems may not allow all flags to be changed.  For
	// example, many file systems will not allow a change from read-write to
	// read-only.
	MntUpdate = C.MNT_UPDATE

	// The flag MNT_RELOAD causes the vfs subsystem to update its data structures 
	// pertaining to the specified already mounted file system.
	MntReload = C.MNT_RELOAD
)

