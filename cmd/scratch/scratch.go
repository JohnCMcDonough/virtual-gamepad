//go:build linux
// +build linux

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/udev"
	"github.com/pilebones/go-udev/netlink"
)

func main() {
	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var mode netlink.Mode

	if os.Getenv("KERNEL_MODE") == "true" {
		mode = netlink.KernelEvent
		log.Default().Printf("Running in Kernel mode")
	} else {
		mode = netlink.UdevEvent
		log.Default().Printf("Running in Udev mode")

	}

	udevConn, err := udev.NewUdevNetlink(mode)

	if err != nil {
		log.Fatalf("Unable to connect to kernel udev netlink %v", err)
	}

	eventChannel := make(chan netlink.UEvent, 1)
	errorChannel := make(chan error, 1)
	monitorCancel := udevConn.Monitor(eventChannel, errorChannel, nil)

	go func() {
		for {
			select {
			case event := <-eventChannel:
				udev.PrintUEvent(event)
			case <-sigs:
				done <- true
				close(monitorCancel)
				return
			}
		}
	}()

	// evt := netlink.UEvent{Action: netlink.ADD, KObj: "/dev/input/js0", Env: map[string]string{
	// 	"ACTION":            "add",
	// 	"DEVNAME":           "/dev/input/js0",
	// 	"DEVPATH":           "/devices/virtual/input/input343/js0",
	// 	"ID_INPUT":          "1",
	// 	"ID_INPUT_JOYSTICK": "1",
	// 	"MAJOR":             "13",
	// 	"MINOR":             "76",
	// 	"SEQNUM":            "13848",
	// 	"SUBSYSTEM":         "input",
	// 	"USEC_INITIALIZED":  "1486458286858",
	// }}

	// err = udevConn.Write(evt)

	if err != nil {
		log.Default().Printf("Failed to write udev action: %v", err)
	}

	<-done

}
