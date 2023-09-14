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

const MAX_GAMEPADS = 15

var stringShouldCreateGamepads, _ = os.LookupEnv("CREATE_GAMEPADS")
var shouldCreateGamepads = stringShouldCreateGamepads != "false"

type GamepadHub struct {
	gamepads [MAX_GAMEPADS]uinput.Gamepad
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
	if shouldCreateGamepads {
		for i := 0; i < len(h.gamepads); i++ {
			gamepad, err := uinput.CreateGamepad("/dev/uinput", []byte(fmt.Sprintf("Gamepad %d", i+1)), 0x03, 0x03)
			if err != nil {
				h.Log.Error().Err(err).Msg("Failed to create gamepad... aborting")
				for j := i - 1; j >= 0; j-- {
					h.Log.Error().Msg(fmt.Sprintf("Closing gamepad %d", j))
					h.gamepads[i].Close()
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

func (h *GamepadHub) Stop() error {
	h.Log.Info().Msg("GamepadHub Stopped")

	return nil
}

var topicRegex = regexp.MustCompile(`^/gamepad/(\d+)/([^/]+)$`)

func setButtonState(gamepad uinput.Gamepad, key int, pressed bool) {
	if pressed {
		gamepad.ButtonDown(key)
	} else {
		gamepad.ButtonUp(key)
	}
}

// /gamepad/<id>/state
func (h *GamepadHub) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info().Msg("OnPublish: 0x" + hex.EncodeToString(pk.Payload))

	// evaluate regex and find all matches
	topicComps := topicRegex.FindStringSubmatch(pk.TopicName)
	h.Log.Info().Msg(fmt.Sprintf("Handling message for Gamepad ID %s with action %s", topicComps[1], topicComps[2]))

	if topicComps == nil {
		return pk, nil // the message wasn't meant for us
	}

	// get the gamepad id as an int
	gamepadID, err := strconv.Atoi(topicComps[1])
	if err != nil || gamepadID < 0 || gamepadID > MAX_GAMEPADS {
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
			setButtonState(gamepad, uinput.ButtonNorth, gamepadState.ButtonNorth)
			setButtonState(gamepad, uinput.ButtonSouth, gamepadState.ButtonSouth)
			setButtonState(gamepad, uinput.ButtonWest, gamepadState.ButtonWest)
			setButtonState(gamepad, uinput.ButtonEast, gamepadState.ButtonEast)
			setButtonState(gamepad, uinput.ButtonBumperLeft, gamepadState.ButtonBumperLeft)
			setButtonState(gamepad, uinput.ButtonBumperRight, gamepadState.ButtonBumperRight)
			setButtonState(gamepad, uinput.ButtonThumbLeft, gamepadState.ButtonThumbLeft)
			setButtonState(gamepad, uinput.ButtonThumbRight, gamepadState.ButtonThumbRight)
			setButtonState(gamepad, uinput.ButtonSelect, gamepadState.ButtonSelect)
			setButtonState(gamepad, uinput.ButtonStart, gamepadState.ButtonStart)
			setButtonState(gamepad, uinput.ButtonDpadUp, gamepadState.ButtonDpadUp)
			setButtonState(gamepad, uinput.ButtonDpadDown, gamepadState.ButtonDpadDown)
			setButtonState(gamepad, uinput.ButtonDpadLeft, gamepadState.ButtonDpadLeft)
			setButtonState(gamepad, uinput.ButtonDpadRight, gamepadState.ButtonDpadRight)
			setButtonState(gamepad, uinput.ButtonMode, gamepadState.ButtonMode)

			gamepad.LeftStickMove(gamepadState.AxisLeftX, gamepadState.AxisLeftY)
			gamepad.RightStickMove(gamepadState.AxisRightX, gamepadState.AxisRightY)
			// todo use analog triggers later
			setButtonState(gamepad, uinput.ButtonTriggerLeft, gamepadState.AxisLeftTrigger > 0.5)
			setButtonState(gamepad, uinput.ButtonTriggerRight, gamepadState.AxisRightTrigger > 0.5)
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
