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
	"net"
	"os/exec"
	"strconv"
	"syscall"
)

// Jail provides a wrapper around a single jail's metadata. Older versions of
// FreeBSD represented jails as a simple struct, however there's been so much
// added to them that you have to use an iovec to access most functionality.
// As such, there's not much corresponding documentation -- the best thing is
// jail(8) which provides the name and description of most iovec keys.
//
// Jail exposes some of the jail functionality, but not all (since there's a lot
// of it, and most of it is off the beaten path). One big difference is that
// all functionality is for persistent jails -- transient jails (e.g., a jail
// that exists until the spawned process terminates) are not supported. This is
// because that's the direction the jails have been moving the past several years.
type Jail struct {
	jid    int
	parent int
	name   string

	hostname string

	path     string
	cpusetId int

	addrs []net.IP
}

// NewJail allocates a new persistent jail with the specified name/path. Name
// should be a unique identifier on this system for the jail; if empty, it will
// be set by the OS for you (the values generated are simply a string
// representation of the jail ID). Path should be valid path that exists.
//
// Nil will not be returned without error.
func NewJail(name string, path string) (*Jail, error) {
	jpps := jailParamList{}
	defer jpps.release()

	params := map[string]interface{}{
		"name":          name,
		"host.hostname": name,
		"path":          path,
		"persist":       true,
	}

	jpps.bindParameters(params)

	jid, er := C.jailparam_set(&jpps.params[0], jpps.numParams(), C.JAIL_CREATE)
	if er != nil {
		return nil, er
	}

	jail := &Jail{
		jid: int(jid),
	}

	if er := jail.Refresh(); er != nil {
		return nil, er
	}

	return jail, nil
}

// EnumerateJails returns a list of all Jails currently on the system, created by
// this package or otherwise (and may contain ephemeral jails). The returned slice
// should be considered passed-by-value, that is, any changes made to the underlying
// data will not be reflected on the returned objects. Calling Refresh will sync
// the Jail instances with the current values.
func EnumerateJails() (jails []Jail, er error) {
	jpps := jailParamList{}
	defer jpps.release()

	var lastjid C.int = 0

	if er := jpps.bindParameter("lastjid", &lastjid); er != nil {
		return nil, er
	}

	for lastjid >= 0 {
		lastjid, er = C.jailparam_get(&jpps.params[0], jpps.numParams(), 0)

		if er == nil {
			jails = append(jails, Jail{jid: int(lastjid)})

		} else if er != syscall.ENOENT {
			return nil, er
		}
	}

	for i := range jails {
		if er := jails[i].Refresh(); er != nil {
			return nil, er
		}
	}

	return jails, nil
}

// Refresh synchronizes the cached values for the Jail fields with the actual
// current state by querying the OS. In general, you probably shouldn't be
// touching other people's jails, so they shouldn't be changing under you.
func (j *Jail) Refresh() error {
	jpps := jailParamList{}
	defer jpps.release()

	if er := jpps.bindParameter("jid", &j.jid); er != nil {
		return er
	}

	if er := jpps.bindOutputs("parent", "name", "host.hostname", "path", "cpuset.id", "ip4.addr"); er != nil {
		return er
	}

	_, er := C.jailparam_get(&jpps.params[0], jpps.numParams(), 0)
	if er != nil {
		return er
	}

	if er := jpps.grabOutput("name", &j.name); er != nil {
		return er
	}

	if er := jpps.grabOutput("host.hostname", &j.hostname); er != nil {
		return er
	}

	if er := jpps.grabOutput("path", &j.path); er != nil {
		return er
	}

	if er := jpps.grabOutput("ip4.addr", &j.addrs); er != nil {
		return er
	}

	return nil
}

// Jid returns the OS-assigned jail ID for this Jail. This ID is not stable, but
// is generally monotonically incrementing.
func (j *Jail) Jid() int {
	return j.jid
}

// Parent returns the jail ID for Jail's parent; if the Jail's parent is not a jail,
// it returns 0.
func (j *Jail) Parent() int {
	return j.parent
}

// Name returns the unique identifier used to create the Jail.
func (j *Jail) Name() string {
	return j.name
}

// Hostname returns the current hostname of the Jail. When first created, the
// hostname is the same as the Jail name.
func (j *Jail) Hostname() string {
	return j.hostname
}

// SetHostname changes the hostname for the jail.
func (j *Jail) SetHostname(hostname string) error {
	jpps := jailParamList{}
	defer jpps.release()

	params := map[string]interface{}{
		"jid":           &j.jid,
		"host.hostname": hostname,
	}

	if er := jpps.bindParameters(params); er != nil {
		return er
	}

	if _, er := C.jailparam_set(&jpps.params[0], jpps.numParams(), C.JAIL_UPDATE); er != nil {
		return er
	}

	j.hostname = hostname

	return nil
}

// Path returns the path specified when the jail was created. 
func (j *Jail) Path() string {
	return j.path
}

// CpusetId returns the cpuset id assigned to this jail, if any.
//
// XXX: Should have a separate package that exposes this functionality.
func (j *Jail) CpusetId() int {
	return j.cpusetId
}

// SetCpusetId assigns a cpuset id to the jail. 
//
// XXX: Need cpuset functionality.
func (j *Jail) SetCpusetId(id int) error {
	jpps := jailParamList{}
	defer jpps.release()

	params := map[string]interface{}{
		"jid":       &j.jid,
		"cpuset.id": id,
	}

	if er := jpps.bindParameters(params); er != nil {
		return er
	}

	if _, er := C.jailparam_set(&jpps.params[0], jpps.numParams(), C.JAIL_UPDATE); er != nil {
		return er
	}

	j.cpusetId = id

	return nil
}

// Exec wraps a command with `jexec` such that it can be spawned within the
// receiving jail. Since much of the Command functionality is exposed via
// settable fields before starting the command (e.g., env, stdin/out, etc)
// starting the command is left to the caller. Note that by default, the
// command is started with uid/gid=0 (though jailed), and must setuid away 
// the privledges.
func (j *Jail) Exec(cmd string, args ...string) *exec.Cmd {
	/* This is kind of a hack, but implementing this in Go is a non-trivial amount of code;
	 * basically, we need to fork+jail_attach+execvp, but forking has ... strange implications.
	 * Instead, we rely on another binary to leverage Go's built-in Exec stuff. */
	newArgs := make([]string, len(args)+2)
	newArgs[0] = strconv.FormatInt(int64(j.jid), 10)
	newArgs[1] = cmd
	copy(newArgs[2:], args)

	return exec.Command("jexec", newArgs...)
}

// Attach locks the current process in the specified jail.
func (j *Jail) Attach() error {
	_, er := C.jail_attach(C.int(j.jid))
	return er
}

// Destroy shuts down the jail. This is very harsh -- it is equivalent to a 
// `killall -9 *` in the jail, and could result in bad things happening if
// there are things that aren't yet shut down cleanly.
//
// XXX: Make a nicer version that sends SIGTERM, waits, then SIGKILL.
func (j *Jail) Destroy() error {
	_, er := C.jail_remove(C.int(j.jid))
	return er
}
