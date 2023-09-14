package uevent

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// NETLINK_KOBJECT_UEVENT is the socket protocol for kernel uevent,
// see /usr/include/linux/netlink.h
// Reader implements reading uevents from an AF_NETLINK socket.
type UEventSocket struct {
	fd     int // the file descriptor of the socket.
	closed bool
}

var _ io.ReadCloser = (*UEventSocket)(nil)

// Read reads from the underlying netlink socket.
// Trying to read from a closed reader return io.EOF.
func (r *UEventSocket) Read(p []byte) (n int, err error) {
	n, err = syscall.Read(r.fd, p)
	// If the underlying socket has been closed with Reader.Close()
	// syscall.Read() returns a -1 and an EBADF error.
	// This Read() function is called by bufio.Reader.ReadString() that
	// panics if a negative number of read bytes is returned.
	// Since the EBADF errors could either mean that the file
	// descriptor is not valid or not open for reading we keep track
	// if it's actually closed or not and return an io.EOF.
	if r.closed {
		return 0, io.EOF
	}
	return
}

// Write writes from the underlying netlink socket.
// Trying to write to a closed reader return io.EOF.
func (r *UEventSocket) Write(p []byte) (n int, err error) {
	n, err = syscall.Write(r.fd, p)
	// If the underlying socket has been closed with Reader.Close()
	// syscall.Read() returns a -1 and an EBADF error.
	// This Read() function is called by bufio.Reader.ReadString() that
	// panics if a negative number of read bytes is returned.
	// Since the EBADF errors could either mean that the file
	// descriptor is not valid or not open for reading we keep track
	// if it's actually closed or not and return an io.EOF.
	if r.closed {
		return 0, io.EOF
	}
	if n != len(p) {
		return 0, fmt.Errorf("not all bytes were written %v != %v", n, len(p))
	}
	return
}

// Close closes the underlying netlink socket.
func (r *UEventSocket) Close() error {
	if r.closed {
		// Already closed, nothing to do
		return nil
	}
	r.closed = true
	return syscall.Close(r.fd)
}

// NewReader returns a new netlink socket reader.
// It opens a raw AF_NETLINK domain socket using the uevent protocol
// and binds it to the PID of the calling program.
func NewSocket() (io.ReadWriteCloser, error) {
	fd, err := syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_RAW,
		NETLINK_KOBJECT_UEVENT,
	)
	if err != nil {
		return nil, err
	}

	// os/exec does not close existing file descriptors by convention as per
	// https://github.com/golang/go/blob/release-branch.go1.14/src/syscall/exec_linux.go#L483
	// so explicitly mark this file descriptor as close-on-exec to avoid leaking
	// it to child processes accidentally.
	syscall.CloseOnExec(fd)

	nl := syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(os.Getpid()),
		// Groups: 1,
		Groups: 2,
	}

	err = syscall.Bind(fd, &nl)
	return &UEventSocket{fd: fd}, err
}
