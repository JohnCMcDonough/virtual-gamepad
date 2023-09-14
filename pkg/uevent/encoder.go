package uevent

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

// Encoder decodes uevents from a reader.
type Encoder struct {
	w *bufio.Writer
}

// NewEncoder creates an uevent Encoder
// using the given reader to read uevents from.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: bufio.NewWriter(w)}
}

// Decode blocks until the uext uevent happens, decodes and returns it.
// It is meant to be used in a loop.
func (e *Encoder) Encode(evt *UEvent) []byte {
	log.Default().Println("Creating new Encoder")
	var segments [][]byte

	segments = append(segments, []byte(evt.Header))
	segments = append(segments, []byte(fmt.Sprintf("ACTION=%s", evt.Action)))
	segments = append(segments, []byte(fmt.Sprintf("DEVPATH=%s", evt.Devpath)))
	segments = append(segments, []byte(fmt.Sprintf("SUBSYSTEM=%s", evt.Subsystem)))

	if evt.Vars != nil {
		for k, v := range evt.Vars {
			// add additional keys not part of the core uevent message
			segments = append(segments, []byte(fmt.Sprintf("%s=%s", k, v)))
		}
	}

	segments = append(segments, []byte(fmt.Sprintf("SEQNUM=%s", evt.Seqnum)))

	for _, segment := range segments {
		log.Default().Println(string(segment))
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

	// write prefixing "
	var bytes []byte // = []byte("\"")

	for _, seg := range segments {
		bytes = append(bytes, append(seg, byte(0))...)
	}
	// write closing ",
	//bytes = append(bytes, []byte("\",")...)

	log.Default().Printf("%v", bytes)

	return bytes
}
