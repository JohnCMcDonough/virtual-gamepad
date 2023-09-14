package uevent

import (
	"fmt"
)

const NETLINK_KOBJECT_UEVENT = 15

var seqnum int64 = 64562

func NewUEvent(action string, devPath string, vars map[string]string) *UEvent {
	evt := &UEvent{
		Header:    fmt.Sprintf("%s@%s", action, devPath),
		Action:    action,
		Devpath:   devPath,
		Subsystem: "input",
		Seqnum:    fmt.Sprintf("%v", seqnum),
		Vars:      vars,
	}

	seqnum += 1

	return evt
}

/// a message looks like this
// "add@/class/input/input9/mouse2\0    // message
// ACTION=add\0                         // action type
// DEVPATH=/class/input/input9/mouse2\0 // path in /sys
// SUBSYSTEM=input\0                    // subsystem (class)
// SEQNUM=1064\0                        // sequence number
// PHYSDEVPATH=/devices/pci0000:00/0000:00:1d.1/usb2/2­2/2­2:1.0\0  // device path in /sys
// PHYSDEVBUS=usb\0       // bus
// PHYSDEVDRIVER=usbhid\0 // driver
// MAJOR=13\0             // major number
// MINOR=34\0",           // minor number

// UEvent represents a single uevent.
type UEvent struct {
	Header string

	// default uevent variables as per kobject_uevent.c
	Action    string
	Devpath   string
	Subsystem string
	Seqnum    string

	// A key/value map of all variables
	Vars map[string]string
}
