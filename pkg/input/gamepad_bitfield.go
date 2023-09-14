package input

import (
	"encoding/binary"
	"fmt"
	"math"
)

type GamepadBitfield struct {
	// Button 1
	ButtonNorth,
	ButtonSouth,
	ButtonWest,
	ButtonEast,

	ButtonBumperLeft,
	ButtonBumperRight,
	ButtonThumbLeft,
	ButtonThumbRight,

	ButtonSelect,
	ButtonStart,

	ButtonDpadUp,
	ButtonDpadDown,
	ButtonDpadLeft,
	ButtonDpadRight,

	ButtonMode bool

	// Axis 1
	AxisLeftX,
	AxisLeftY,
	// Axis 2
	AxisRightX,
	AxisRightY,

	AxisLeftTrigger,
	AxisRightTrigger float32
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *GamepadBitfield) UnmarshalBinary(data []byte) error {
	if len(data) != 26 {
		return fmt.Errorf("invalid payload size %d should be 26 bytes", len(data))
	}

	// extract the buttons from a packed bitfield
	b.ButtonNorth = data[0]&(1<<0) != 0
	b.ButtonSouth = data[0]&(1<<1) != 0
	b.ButtonWest = data[0]&(1<<2) != 0
	b.ButtonEast = data[0]&(1<<3) != 0

	// extract the bumpers from a packed bitfield
	b.ButtonBumperLeft = data[0]&(1<<4) != 0
	b.ButtonBumperRight = data[0]&(1<<5) != 0

	// extract the thumbstick buttons from a packed bitfield
	b.ButtonThumbLeft = data[0]&(1<<6) != 0
	b.ButtonThumbRight = data[0]&(1<<7) != 0

	// extract the select and start buttons from a packed bitfield
	b.ButtonSelect = data[1]&(1<<0) != 0
	b.ButtonStart = data[1]&(1<<1) != 0

	// extract the dpad buttons from a packed bitfield
	b.ButtonDpadUp = data[1]&(1<<2) != 0
	b.ButtonDpadDown = data[1]&(1<<3) != 0
	b.ButtonDpadLeft = data[1]&(1<<4) != 0
	b.ButtonDpadRight = data[1]&(1<<5) != 0

	// extract button mode
	b.ButtonMode = data[1]&(1<<6) != 0

	// read last bit as a version flag
	// _ := data[1]&(1<<7) != 0

	// extract left thumbstick in little-endian format
	b.AxisLeftX = math.Float32frombits(binary.LittleEndian.Uint32(data[2:6]))
	b.AxisLeftY = math.Float32frombits(binary.LittleEndian.Uint32(data[6:10]))

	// extract right thumbstick in little-endian format
	b.AxisRightX = math.Float32frombits(binary.LittleEndian.Uint32(data[10:14]))
	b.AxisRightY = math.Float32frombits(binary.LittleEndian.Uint32(data[14:18]))

	// extract left and right trigger in little-endian format
	b.AxisLeftTrigger = math.Float32frombits(binary.LittleEndian.Uint32(data[18:22]))
	b.AxisRightTrigger = math.Float32frombits(binary.LittleEndian.Uint32(data[22:26]))

	return nil
}
