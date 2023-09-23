package input

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/gamepad"
	"github.com/JohnCMcDonough/virtual-gamepad/pkg/logger"
	"github.com/JohnCMcDonough/virtual-gamepad/pkg/udev"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/rs/zerolog"
)

var MAX_GAMEPADS int = 4

var stringShouldCreateGamepads, _ = os.LookupEnv("CREATE_GAMEPADS")
var shouldCreateGamepads = stringShouldCreateGamepads != "false"

type GamepadHub struct {
	gamepads []*gamepad.VirtualGamepad
	VendorId uint16
	DeviceId uint16
	lock     sync.Mutex
	l        zerolog.Logger
	udev     *udev.UDev
	mqtt.HookBase
}

func (h *GamepadHub) ID() string {
	return "GamepadHub"
}

func (h *GamepadHub) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
		mqtt.OnConnect,
	}, []byte{b})
}

// OnConnect is called when a new client connects.
func (h *GamepadHub) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.l.Info().Msgf("Received mqtt connection from %v", cl.Net.Conn.RemoteAddr().String())
	return nil
}

func (h *GamepadHub) GetGamepads() []*gamepad.VirtualGamepad {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.gamepads[:]
}

func (h *GamepadHub) Init(config any) error {
	h.l.Info().Msg("GamepadHub Initialized")

	var err error

	if err != nil {
		h.l.Err(err).Msg("There was an error opening the udev socket")
		return err
	}

	vendorIdString := os.Getenv("VIRT_DEVICE_ID")
	deviceIdString := os.Getenv("VIRT_DEVICE_VENDOR")

	if vendorIdString != "" {
		i, err := strconv.ParseInt(vendorIdString, 10, 16)
		if err != nil {
			h.VendorId = uint16(i)
		}
	} else {
		h.VendorId = 0x045E
	}
	if deviceIdString != "" {
		i, err := strconv.ParseInt(deviceIdString, 10, 16)
		if err != nil {
			h.DeviceId = uint16(i)
		}
	} else {
		h.DeviceId = 0x02D1
	}

	if shouldCreateGamepads {
		for i := 0; i < len(h.gamepads); i++ {
			if gamepad, err := gamepad.CreateVirtualGamepad(h.udev, i, int16(h.VendorId), int16(h.DeviceId)); err != nil {
				h.l.Error().Err(err).Msgf("Failed to create Gamepad %v", i)
			} else {
				h.gamepads[i] = gamepad
			}
		}
	} else {
		h.l.Warn().Msg("Skipping gamepad init due to CREATE_GAMEPADS = false")
	}

	return nil
}

func (h *GamepadHub) Stop() error {
	h.l.Info().Msg("GamepadHub Stopped")
	return nil
}

var topicRegex = regexp.MustCompile(`^/gamepad/(\d+)/([^/]+)$`)

// /gamepad/<id>/state
func (h *GamepadHub) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.l.Trace().MsgFunc(func() string { return "OnPublish: 0x" + hex.EncodeToString(pk.Payload) })

	// evaluate regex and find all matches
	topicComps := topicRegex.FindStringSubmatch(pk.TopicName)

	if topicComps == nil {
		return pk, nil // the message wasn't meant for us
	}

	h.l.Trace().MsgFunc(func() string {
		return fmt.Sprintf("Handling message for Gamepad ID %s with action %s", topicComps[1], topicComps[2])
	})

	// get the gamepad id as an int
	gamepadID, err := strconv.Atoi(topicComps[1])
	if err != nil || gamepadID < 0 || gamepadID >= MAX_GAMEPADS {
		h.l.Warn().Msg(fmt.Sprintf("Received event for gamepad that doesn't exist %s", topicComps[1]))
		return pk, nil
	}

	// get the actual pad
	pad := h.gamepads[gamepadID]
	gamepadAction := topicComps[2]

	if gamepadAction == "state" {
		// unmarshal the payload into a bitfield
		var gamepadState gamepad.GamepadBitfield
		err := gamepadState.UnmarshalBinary(pk.Payload)
		if err != nil {
			h.l.Err(err).Msg("Invalid payload")
			return pk, nil
		}
		if shouldCreateGamepads {
			pad.SendInput(gamepadState)
		} else {
			h.l.Info().Msg(fmt.Sprintf("Pad %d state — %+v", gamepadID, gamepadState))
		}

	} else {
		// do nothing
		h.l.Debug().Msgf("Received unknown action for gamepad — %v", gamepadAction)
	}

	return pk, nil
}

func (h *GamepadHub) Close() error {
	errs := []error{}
	for _, vg := range h.GetGamepads() {
		errs = append(errs, vg.Close())
	}
	errs = append(errs, h.udev.Close())
	return errors.Join(errs...)
}

func NewGamepadHub() *GamepadHub {
	hub := new(GamepadHub)
	hub.gamepads = make([]*gamepad.VirtualGamepad, MAX_GAMEPADS)
	hub.l = logger.CreateLogger(map[string]string{
		"Component": "GamepadHub",
	})
	if dev, err := udev.CreateUDev(); err != nil {
		hub.l.Error().Err(err).Msg("Failed to create a connection to udev... aborting")
		os.Exit(1)
	} else {
		hub.udev = dev
	}

	return hub
}
