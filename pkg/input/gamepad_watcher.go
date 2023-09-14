package input

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/uevent"
	"github.com/pilebones/go-udev/netlink"
	"golang.org/x/sys/unix"
)

type GamepadWatcher struct {
	Hub        *GamepadHub
	kSocket    *uevent.UdevEventConnection
	udevSocket *uevent.UdevEventConnection
}

func NewGamepadWatcher(ctx context.Context, hub *GamepadHub) *GamepadWatcher {
	watcher := &GamepadWatcher{Hub: hub}

	return watcher
}

// KERNEL EVENT
//
//	Action: add
//	KObj: /devices/virtual/input/input378/js2
//	Env: map[
//			ACTION:add
//			DEVNAME:input/js2
//			DEVPATH:/devices/virtual/input/input378/js2
//			MAJOR:13
//			MINOR:2
//			SEQNUM:65284
//			SUBSYSTEM:input
//	]
//
// UDEV EVENT
//
//	Action: add
//	KObj: /devices/virtual/input/input358/js2
//	Env: map[
//		ACTION:add
//		DEVNAME:/dev/input/js2
//		DEVPATH:/devices/virtual/input/input358/js2
//		ID_INPUT:1
//		ID_INPUT_JOYSTICK:1
//		MAJOR:13
//		MINOR:2
//		SEQNUM:65164
//		SUBSYSTEM:input
//		USEC_INITIALIZED:1492905162910
//	]
func (w *GamepadWatcher) ProcessEvent(ctx context.Context, evt netlink.UEvent) error {
	time.Sleep(time.Second * 1)
	if evt.Action != netlink.ADD {
		return nil
	}

	if _, ok := evt.Env["MAJOR"]; !ok {
		log.Default().Println("Ignoring event without major version")
		return nil
	}

	log.Default().Println("Processing Event:")
	uevent.PrintUEvent(evt)

	major, err := strconv.Atoi(evt.Env["MAJOR"])
	if err != nil {
		return err
	}
	minor, err := strconv.Atoi(evt.Env["MINOR"])
	if err != nil {
		return err
	}
	devPath := "/sys" + evt.Env["DEVPATH"]

	for i, gp := range w.Hub.GetGamepads() {
		if gp == nil {
			continue
		}
		path, err := gp.FetchSyspath()
		path += "/js"
		if err != nil {
			log.Default().Printf("There was an error fetching Syspath: %v", err)
			continue
		}
		if !strings.HasPrefix(devPath, path) {
			log.Default().Printf("Failed to match prefix %v to %v", path, devPath)
			continue
		}
		log.Default().Printf("Found match %v on gamepad %v\n", devPath, path)

		devName := fmt.Sprintf("/dev/input/js%d", i)
		dev := unix.Mkdev(uint32(major), uint32(minor))
		err = syscall.Mknod(devName, syscall.S_IFCHR|0o666, int(dev))
		if err != nil {
			return fmt.Errorf("failed to mknod: %v", err)
		}

		uEvt := netlink.UEvent{
			Action: evt.Action,
			KObj:   evt.KObj,
			Env:    make(map[string]string),
		}
		for k, v := range evt.Env {
			uEvt.Env[k] = v
		}
		uEvt.Env["DEVNAME"] = devName
		uEvt.Env["ID_INPUT"] = "1"
		uEvt.Env["ID_INPUT_JOYSTICK"] = "1"
		uEvt.Env["ID_SERIAL"] = fmt.Sprintf("%d", i)
		uEvt.Env["USEC_INITIALIZED"] = fmt.Sprintf("%d", time.Now().UnixNano()/1000)

		err = w.udevSocket.Write(uEvt)

		if err != nil {
			return fmt.Errorf("failed to write to udev socket: %v", err)
		}

		return nil
	}

	return nil
}

func (w *GamepadWatcher) Watch(ctx context.Context) error {
	logger := log.Default()

	kSocket, err := uevent.NetUdevNetlink(netlink.KernelEvent)
	if err != nil {
		return err
	}
	w.kSocket = kSocket

	udevSocket, err := uevent.NetUdevNetlink(netlink.UdevEvent)
	if err != nil {
		return err
	}
	w.udevSocket = udevSocket

	eventChannel := make(chan netlink.UEvent, 1)
	errorChannel := make(chan error, 1)
	monitorCancel := w.kSocket.Monitor(eventChannel, errorChannel, nil)
	go func() {
		defer kSocket.Close()
		defer udevSocket.Close()
		defer close(monitorCancel)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChannel:
				err := w.ProcessEvent(ctx, event)
				if err != nil {
					logger.Printf("%v\n", err)
				}
			case err := <-errorChannel:
				logger.Printf("%v\n", err)
			}
		}
	}()

	return nil
}
