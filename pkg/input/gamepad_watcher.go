package input

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"syscall"
// 	"time"

// 	"github.com/JohnCMcDonough/virtual-gamepad/pkg/udev"
// 	"github.com/pilebones/go-udev/netlink"
// 	"golang.org/x/sys/unix"
// )

// type GamepadWatcher struct {
// 	Hub        *GamepadHub
// 	kSocket    *udev.UDevNetlinkConnection
// 	udevSocket *udev.UDevNetlinkConnection

// 	mknodPaths []string
// }

// func NewGamepadWatcher(ctx context.Context, hub *GamepadHub) *GamepadWatcher {
// 	watcher := &GamepadWatcher{Hub: hub}

// 	return watcher
// }

// // KERNEL EVENT
// //
// //	Action: add
// //	KObj: /devices/virtual/input/input378/js2
// //	Env: map[
// //			ACTION:add
// //			DEVNAME:input/js2
// //			DEVPATH:/devices/virtual/input/input378/js2
// //			MAJOR:13
// //			MINOR:2
// //			SEQNUM:65284
// //			SUBSYSTEM:input
// //	]
// //
// // UDEV EVENT
// //
// //	Action: add
// //	KObj: /devices/virtual/input/input358/js2
// //	Env: map[
// //		ACTION:add
// //		DEVNAME:/dev/input/js2
// //		DEVPATH:/devices/virtual/input/input358/js2
// //		ID_INPUT:1
// //		ID_INPUT_JOYSTICK:1
// //		MAJOR:13
// //		MINOR:2
// //		SEQNUM:65164
// //		SUBSYSTEM:input
// //		USEC_INITIALIZED:1492905162910
// //	]
// func (w *GamepadWatcher) ProcessEvent(ctx context.Context, evt netlink.UEvent) error {
// 	if evt.Action != netlink.ADD {
// 		return nil
// 	}

// 	if _, ok := evt.Env["MAJOR"]; !ok {
// 		log.Default().Println("Ignoring event without major version")
// 		return nil
// 	}

// 	log.Default().Println("Processing Event:")
// 	udev.PrintUEvent(evt)

// 	major, err := strconv.Atoi(evt.Env["MAJOR"])
// 	if err != nil {
// 		return err
// 	}
// 	minor, err := strconv.Atoi(evt.Env["MINOR"])
// 	if err != nil {
// 		return err
// 	}
// 	devPath := "/sys" + evt.Env["DEVPATH"]

// 	for i, gp := range w.Hub.GetGamepads() {
// 		if gp == nil {
// 			continue
// 		}
// 		path, err := gp.FetchSyspath()
// 		if err != nil {
// 			log.Default().Printf("There was an error fetching Syspath: %v", err)
// 			continue
// 		}
// 		if !strings.HasPrefix(devPath, path) {
// 			log.Default().Printf("Failed to match prefix %v to %v", path, devPath)
// 			continue
// 		}
// 		log.Default().Printf("Found match %v on gamepad %v\n", devPath, path)

// 		var devName string

// 		if strings.HasPrefix(devPath, path+"/js") {
// 			log.Default().Printf("Using older joystick api (js0-js3)")
// 			devName = fmt.Sprintf("/dev/input/js%d", i)
// 		} else if strings.HasPrefix(devPath, path+"/event") {
// 			log.Default().Printf("Using newer evdev api (event0-9999)")
// 			devName = "/dev/" + evt.Env["DEVNAME"]
// 		} else {
// 			log.Default().Printf("Unknown device API")
// 			return nil
// 		}

// 		if err, _ := os.Stat(devName); err != nil {
// 			// file exists
// 			if err2 := os.Remove(devName); err2 != nil {
// 				return fmt.Errorf("failed to remove existing binding for %s: %v", devName, err)
// 			}
// 		}

// 		dev := unix.Mkdev(uint32(major), uint32(minor))
// 		err = syscall.Mknod(devName, syscall.S_IFCHR|0o777, int(dev))
// 		log.Default().Printf("Running mknod %v with %o %d:%d", devName, syscall.S_IFCHR|0o777, major, minor)
// 		if err != nil {
// 			return fmt.Errorf("failed to mknod: %v", err)
// 		}

// 		if err := os.Chmod(devName, syscall.S_IFCHR|0o777); err != nil {
// 			return fmt.Errorf("failed to set correct permissions on file")
// 		}

// 		w.mknodPaths = append(w.mknodPaths, devName)

// 		uEvt := netlink.UEvent{
// 			Action: evt.Action,
// 			KObj:   evt.KObj,
// 			Env:    make(map[string]string),
// 		}
// 		for k, v := range evt.Env {
// 			uEvt.Env[k] = v
// 		}
// 		uEvt.Env["DEVNAME"] = devName
// 		uEvt.Env["ID_INPUT"] = "1"
// 		uEvt.Env["ID_INPUT_JOYSTICK"] = "1"
// 		uEvt.Env["ID_SERIAL"] = fmt.Sprintf("%d", i)
// 		uEvt.Env["USEC_INITIALIZED"] = fmt.Sprintf("%d", time.Now().UnixNano()/1000)

// 		err = w.udevSocket.Write(uEvt)

// 		if err != nil {
// 			return fmt.Errorf("failed to write to udev socket: %v", err)
// 		}

// 		return nil
// 	}

// 	return nil
// }

// func (w *GamepadWatcher) cleanup() {
// 	for _, path := range w.mknodPaths {
// 		log.Default().Printf("Cleaning up device path %v\n", path)
// 		err := unix.Unlink(path)
// 		if err != nil {
// 			log.Default().Printf("Failed to delete device path %v\n", path)
// 		}
// 	}
// }

// func (w *GamepadWatcher) Watch(ctx context.Context) error {
// 	logger := log.Default()

// 	kSocket, err := udev.NewUdevNetlink(netlink.KernelEvent)
// 	if err != nil {
// 		return err
// 	}
// 	w.kSocket = kSocket

// 	udevSocket, err := udev.NewUdevNetlink(netlink.UdevEvent)
// 	if err != nil {
// 		return err
// 	}
// 	w.udevSocket = udevSocket

// 	eventChannel := make(chan netlink.UEvent, 1)
// 	errorChannel := make(chan error, 1)
// 	monitorCancel := w.kSocket.Monitor(eventChannel, errorChannel, nil)
// 	go func() {
// 		defer kSocket.Close()
// 		defer udevSocket.Close()
// 		defer close(monitorCancel)

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				w.cleanup()
// 				return
// 			case event := <-eventChannel:
// 				err := w.ProcessEvent(ctx, event)
// 				if err != nil {
// 					logger.Printf("%v\n", err)
// 				}
// 			case err := <-errorChannel:
// 				logger.Printf("%v\n", err)
// 			}
// 		}
// 	}()

// 	return nil
// }
