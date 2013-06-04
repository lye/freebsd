# FreeBSD Go Bindings

`package freebsd` provides Go bindings to some of the more uncommon (but very useful!) features of FreeBSD. All functionality should be considered untested, APIs should be considered unstable (e.g., fork before using). Check the docs for the following sub-packages:

 * [`package fs`](http://godoc.org/github.com/lye/freebsd/fs) provides bindings to nmount and statfs, allowing filesystem manipulation.
 * [`package jail`](http://godoc.org/github.com/lye/freebsd/jail) provides an interface for creating and managing jails.
 * [`package md`](http://godoc.org/github.com/lye/freebsd/md) provides an interface to malloc/vnode/swap-backed `md` devices.
 * [`package netif`](http://godoc.org/github.com/lye/freebsd/netif) will eventually provide access to the a system's network interfaces, but right now just does address enumeration.
