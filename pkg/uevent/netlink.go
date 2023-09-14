package uevent

import (
	"io"
	"syscall"

	netlink "github.com/pilebones/go-udev/netlink"
)

type UdevEventConnection struct {
	closed bool
	netlink.UEventConn
}

func (c *UdevEventConnection) Write(event netlink.UEvent) (err error) {
	data := event.Bytes()
	err = syscall.Sendto(c.Fd, data, 0, &c.Addr)
	// If the underlying socket has been closed with Reader.Close()
	// syscall.Read() returns a -1 and an EBADF error.
	// This Read() function is called by bufio.Reader.ReadString() that
	// panics if a negative number of read bytes is returned.
	// Since the EBADF errors could either mean that the file
	// descriptor is not valid or not open for reading we keep track
	// if it's actually closed or not and return an io.EOF.
	if c.closed {
		return io.EOF
	}
	return
}

func (c *UdevEventConnection) Close() error {
	if c.closed {
		// Already closed, nothing to do
		return nil
	}
	c.closed = true
	return syscall.Close(c.Fd)
}

func NetUdevNetlink(mode netlink.Mode) (*UdevEventConnection, error) {
	conn := &UdevEventConnection{}
	err := conn.Connect(mode)
	return conn, err
}
