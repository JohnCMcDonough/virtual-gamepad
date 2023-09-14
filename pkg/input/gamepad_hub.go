//go:build linux
// +build linux

package input

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/bendahl/uinput"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

const MAX_GAMEPADS = 4

var stringShouldCreateGamepads, _ = os.LookupEnv("CREATE_GAMEPADS")
var shouldCreateGamepads = stringShouldCreateGamepads != "false"

type GamepadHub struct {
	gamepads [MAX_GAMEPADS]uinput.Gamepad
	vendorId uint16
	deviceId uint16
	// sock     io.ReadWriteCloser
	// enc      *uevent.Encoder
	mqtt.HookBase
}

func (h *GamepadHub) ID() string {
	return "GamepadHub"
}

func (h *GamepadHub) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *GamepadHub) Init(config any) error {
	h.Log.Info().Msg("GamepadHub Initialized")

	var err error

	// h.sock, err = uevent.NewSocket()

	if err != nil {
		h.Log.Err(err).Msg("There was an error opening the udev socket")
		return err
	}

	// h.enc = uevent.NewEncoder(h.sock)

	vendorIdString := os.Getenv("VIRT_DEVICE_ID")
	deviceIdString := os.Getenv("VIRT_DEVICE_VENDOR")

	if vendorIdString != "" {
		i, err := strconv.ParseInt(vendorIdString, 10, 16)
		if err != nil {
			h.vendorId = uint16(i)
		} else {
			h.vendorId = 3
		}
	}
	if deviceIdString != "" {
		i, err := strconv.ParseInt(deviceIdString, 10, 16)
		if err != nil {
			h.deviceId = uint16(i)
		} else {
			h.deviceId = 4
		}
	}

	if shouldCreateGamepads {
		for i := 0; i < len(h.gamepads); i++ {
			gamepad, err := createGamepad(h, i)
			if err != nil {
				h.Log.Error().Err(err).Msg("Failed to create gamepad... aborting")
				for j := i - 1; j >= 0; j-- {
					h.Log.Error().Msg(fmt.Sprintf("Closing gamepad %d", j))
					closeGamepad(h, i)
				}
				return err
			}
			h.gamepads[i] = gamepad
		}
	} else {
		h.Log.Warn().Msg("Skipping gamepad init due to CREATE_GAMEPADS = false")
	}

	return nil
}

func closeGamepad(h *GamepadHub, i int) error {
	err := h.gamepads[i].Close()

	return err
}

func createGamepad(h *GamepadHub, i int) (uinput.Gamepad, error) {
	gamepad, err := uinput.CreateGamepad("/dev/uinput", []byte(fmt.Sprintf("Gamepad %d", i+1)), h.vendorId, h.deviceId)
	if err != nil {
		return nil, err
	}

	// dev := unix.Mkdev();
	// syscall.Mknod(path, mode, dev)

	// uevent := uevent.NewUEvent("add", "")

	return gamepad, nil
}

func (h *GamepadHub) Stop() error {
	h.Log.Info().Msg("GamepadHub Stopped")

	return nil
}

var topicRegex = regexp.MustCompile(`^/gamepad/(\d+)/([^/]+)$`)

func (h *GamepadHub) setButtonState(gamepad uinput.Gamepad, key int, pressed bool) {
	var err error
	if pressed {
		err = gamepad.ButtonDown(key)
	} else {
		err = gamepad.ButtonUp(key)
	}
	if err != nil {
		h.Log.Err(err).Msg(fmt.Sprintf("Unable to set button state for gamepad %d %v", key, pressed))
	}
}

// /gamepad/<id>/state
func (h *GamepadHub) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info().Msg("OnPublish: 0x" + hex.EncodeToString(pk.Payload))

	// evaluate regex and find all matches
	topicComps := topicRegex.FindStringSubmatch(pk.TopicName)

	if topicComps == nil {
		return pk, nil // the message wasn't meant for us
	}

	h.Log.Info().Msg(fmt.Sprintf("Handling message for Gamepad ID %s with action %s", topicComps[1], topicComps[2]))

	// get the gamepad id as an int
	gamepadID, err := strconv.Atoi(topicComps[1])
	if err != nil || gamepadID < 0 || gamepadID >= MAX_GAMEPADS {
		h.Log.Warn().Msg(fmt.Sprintf("Received event for gamepad that doesn't exist %s", topicComps[1]))
		return pk, nil
	}

	// get the actual gamepad
	gamepad := h.gamepads[gamepadID]
	gamepadAction := topicComps[2]

	if gamepadAction == "state" {
		// unmarshal the payload into a bitfield
		var gamepadState GamepadBitfield
		err := gamepadState.UnmarshalBinary(pk.Payload)
		if err != nil {
			h.Log.Err(err).Msg("Invalid payload")
			return pk, nil
		}
		if shouldCreateGamepads {
			h.setButtonState(gamepad, uinput.ButtonNorth, gamepadState.ButtonNorth)
			h.setButtonState(gamepad, uinput.ButtonSouth, gamepadState.ButtonSouth)
			h.setButtonState(gamepad, uinput.ButtonWest, gamepadState.ButtonWest)
			h.setButtonState(gamepad, uinput.ButtonEast, gamepadState.ButtonEast)
			h.setButtonState(gamepad, uinput.ButtonBumperLeft, gamepadState.ButtonBumperLeft)
			h.setButtonState(gamepad, uinput.ButtonBumperRight, gamepadState.ButtonBumperRight)
			h.setButtonState(gamepad, uinput.ButtonThumbLeft, gamepadState.ButtonThumbLeft)
			h.setButtonState(gamepad, uinput.ButtonThumbRight, gamepadState.ButtonThumbRight)
			h.setButtonState(gamepad, uinput.ButtonSelect, gamepadState.ButtonSelect)
			h.setButtonState(gamepad, uinput.ButtonStart, gamepadState.ButtonStart)
			h.setButtonState(gamepad, uinput.ButtonDpadUp, gamepadState.ButtonDpadUp)
			h.setButtonState(gamepad, uinput.ButtonDpadDown, gamepadState.ButtonDpadDown)
			h.setButtonState(gamepad, uinput.ButtonDpadLeft, gamepadState.ButtonDpadLeft)
			h.setButtonState(gamepad, uinput.ButtonDpadRight, gamepadState.ButtonDpadRight)
			h.setButtonState(gamepad, uinput.ButtonMode, gamepadState.ButtonMode)

			gamepad.LeftStickMove(gamepadState.AxisLeftX, gamepadState.AxisLeftY)
			gamepad.RightStickMove(gamepadState.AxisRightX, gamepadState.AxisRightY)
			// todo use analog triggers later
			h.setButtonState(gamepad, uinput.ButtonTriggerLeft, gamepadState.AxisLeftTrigger > 0.5)
			h.setButtonState(gamepad, uinput.ButtonTriggerRight, gamepadState.AxisRightTrigger > 0.5)
		} else {
			h.Log.Info().Msg(fmt.Sprintf("Pad %d state — %+v", gamepadID, gamepadState))
		}

	} else {
		// do nothing
	}

	return pk, nil
}

func NewGamepadHub() *GamepadHub {
	hub := new(GamepadHub)
	return hub
}
