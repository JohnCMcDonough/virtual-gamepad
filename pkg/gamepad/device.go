package gamepad

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/logger"
	"github.com/JohnCMcDonough/virtual-gamepad/pkg/udev"
	"github.com/pilebones/go-udev/netlink"
	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
)

func touch(name string, perm fs.FileMode) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, perm)
	if err != nil {
		return err
	}
	err = os.Chmod(name, perm)
	if err != nil {
		return err
	}
	return file.Close()
}

type Device struct {
	initTime   int64 // should me usec
	OriginalId int
	Id         int
	KObj       string
	Env        map[string]string
	Major      int16
	Minor      int16
	DevPath    string

	udev *udev.UDev
	l    zerolog.Logger
}

/*
 * This function should be called after all of the public properties of the device have been set.
 */
func (d *Device) Initialize(udev *udev.UDev) {
	d.udev = udev
	// var env []string
	// for k, v := range d.Env {
	// 	env = append(env, fmt.Sprintf("%v=%v", k, v))
	// }
	d.l = logger.CreateLogger(map[string]string{
		"Id":         fmt.Sprint(d.Id),
		"OriginalId": fmt.Sprint(d.OriginalId),
	})
	if err := os.MkdirAll("/run/udev/data", 0o755); err != nil {
		d.l.Error().Err(err).Msg("Failed to create /run/udev/data")
	} else if os.Chmod("/run/udev/data", 0o755); err != nil {
		d.l.Error().Err(err).Msg("Failed to set permissions on /run/udev/data")
	}
	if err := touch("/run/udev/control", 0o755); err != nil {
		d.l.Error().Err(err).Msg("Failed to create /run/udev/control")
	}
	if err := os.RemoveAll(d.DevPath); err != nil {
		d.l.Error().Msgf("Failed to remove existing devpath %s", d.DevPath)
	}

	d.initTime = time.Now().UnixNano() / 1000

	d.MakeDeviceNode()
	d.WriteUDevDatabaseData()
	if err := d.EmitUDevEvent(netlink.ADD); err != nil {
		d.l.Error().Err(err).Msg("Failed to emit udev add event")
	}
	d.l.Debug().Msg("Device created successfully")
}

func (d *Device) GetUDevDBPath() string {
	return fmt.Sprintf("/run/udev/data/c%v:%v", d.Major, d.Minor)
}

func (d *Device) WriteUDevDatabaseData() {
	// Write udev database information
	characterDevicePath := d.GetUDevDBPath()
	data := ""
	data += fmt.Sprintf("I:%v\n", d.initTime)
	data += "E:ID_INPUT=1\n"
	data += "E:ID_INPUT_JOYSTICK=1\n"
	data += "E:ID_SERIAL=noserial\n"
	data += "G:seat"
	data += "G:uaccess"
	data += "Q:seat"
	data += "Q:uaccess"
	data += "V:1"
	if err := os.WriteFile(characterDevicePath, []byte(data), 0o755); err != nil {
		d.l.Error().Err(err).Msgf("Failed to write device database to %s", characterDevicePath)
	}
}

func (d *Device) MakeDeviceNode() {
	devId := unix.Mkdev(uint32(d.Major), uint32(d.Minor))
	if err := unix.Mknod(d.DevPath, syscall.S_IFCHR|0o777, int(devId)); err != nil {
		d.l.Error().Err(err).Msgf("Failed to mknod for device path at %s", d.DevPath)
		return
	}
	if err := os.Chmod(d.DevPath, syscall.S_IFCHR|0o777); err != nil {
		d.l.Error().Err(err).Msgf("Failed to update permissions to be more open at %s", d.DevPath)
	}
}

func (d *Device) EmitUDevEvent(action netlink.KObjAction) error {
	evt := netlink.UEvent{
		Action: action,
		KObj:   d.KObj,
		Env:    d.Env,
	}

	evt.Env["ACTION"] = action.String()
	evt.Env["DEVNAME"] = d.DevPath // overwrite the original devpath with our new one
	evt.Env["SUBSYSTEM"] = "input"
	evt.Env["USEC_INITIALIZED"] = fmt.Sprint(d.initTime)
	evt.Env["ID_INPUT"] = "1"
	evt.Env["ID_INPUT_JOYSTICK"] = "1"
	evt.Env[".INPUT_CLASS"] = "joystick"
	evt.Env["ID_SERIAL"] = "noserial"
	evt.Env["TAGS"] = ":seat:uaccess:"
	evt.Env["CURRENT_TAGS"] = ":seat:uaccess:"

	evtString := strings.Join(strings.Split(evt.String(), "\000"), " ")

	d.l.Info().Msgf("Emitting UDev Event %s", evtString)
	return d.udev.WriteUDevEvent(evt)
}

func (d *Device) Close() error {
	if d.udev == nil {
		return nil
	}
	return errors.Join(
		d.EmitUDevEvent(netlink.REMOVE),
		os.Remove(d.DevPath),
		os.Remove(d.GetUDevDBPath()),
	)
}

func (d *Device) GetSysPath() string {
	return "/sys" + d.KObj
}
