package udev

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pilebones/go-udev/netlink"
	eventemitter "github.com/vansante/go-event-emitter"
)

type UDev struct {
	closed bool
	// internal sockets
	kernelSock *UDevNetlinkConnection
	udevSock   *UDevNetlinkConnection

	// Used to listen to events from udev
	UDevEvents *eventemitter.Emitter
	// Used to listen to events from the Kernel
	KernelEvents *eventemitter.Emitter
}

var logger = log.Default()

// mod with number of seconds in a year. I don't know how high these sequence numbers can get...
var seqNum = time.Now().Unix() % int64((24 * time.Hour * 30 * 12).Seconds())

func CreateUDev() (*UDev, error) {
	var udev = &UDev{}
	if kernelSock, err := NewUdevNetlink(netlink.KernelEvent); err != nil {
		return nil, err
	} else {
		udev.kernelSock = kernelSock
	}
	if udevSock, err := NewUdevNetlink(netlink.UdevEvent); err != nil {
		return nil, errors.Join(err, udev.kernelSock.Close())
	} else {
		udev.udevSock = udevSock
	}

	udev.KernelEvents = eventemitter.NewEmitter(false)
	udev.UDevEvents = eventemitter.NewEmitter(false)

	// run forever and read from uevents
	go func() {
		for !udev.closed {
			evt, err := udev.kernelSock.ReadUEvent()
			if err != nil {
				logger.Printf("Error reading kernel UEvent - %v\n", err)
			} else {
				udev.KernelEvents.EmitEvent(eventemitter.EventType(evt.Action.String()), evt)
			}
		}
	}()

	return udev, nil
}

func (u *UDev) Close() error {
	u.closed = true
	return errors.Join(
		u.kernelSock.Close(),
		u.udevSock.Close(),
	)
}

func (u *UDev) WriteUDevEvent(evt netlink.UEvent) error {
	seqNum++
	evt.Env["SEQNUM"] = fmt.Sprint(seqNum)
	u.UDevEvents.EmitEvent(eventemitter.EventType(evt.Action.String()), evt)
	return u.udevSock.Write(evt)
}

func PrintUEvent(evt netlink.UEvent) {
	logger.Printf("Action: %v", evt.Action)
	logger.Printf("KObj: %v", evt.KObj)
	logger.Printf("Env: %v", evt.Env)
}
