//go:build linux
// +build linux

package input

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantErr   bool
		wantField *GamepadBitfield
	}{
		{
			name:    "Invalid payload size",
			data:    []byte{0, 0},
			wantErr: true,
		},
		{
			name: "Valid payload",
			data: func() []byte {
				data := make([]byte, 26)
				data[0] = 0b10101010
				data[1] = 0b01010101
				binary.LittleEndian.PutUint32(data[2:6], 0x41280000) // 10.5 in float32
				binary.LittleEndian.PutUint32(data[6:10], 0x0)       // 0.0 in float32
				binary.LittleEndian.PutUint32(data[10:14], 0x0)      // 0.0 in float32
				binary.LittleEndian.PutUint32(data[14:18], 0x0)      // 0.0 in float32
				binary.LittleEndian.PutUint32(data[18:22], 0x0)      // 0.0 in float32
				binary.LittleEndian.PutUint32(data[22:26], 0x0)      // 0.0 in float32
				fmt.Printf("Hex value is %s", hex.EncodeToString(data))
				return data
			}(),
			wantErr: false,
			wantField: &GamepadBitfield{
				ButtonNorth:       false,
				ButtonSouth:       true,
				ButtonWest:        false,
				ButtonEast:        true,
				ButtonBumperLeft:  false,
				ButtonBumperRight: true,
				ButtonThumbLeft:   false,
				ButtonThumbRight:  true,
				ButtonSelect:      true,
				ButtonStart:       false,
				ButtonDpadUp:      true,
				ButtonDpadDown:    false,
				ButtonDpadLeft:    true,
				ButtonDpadRight:   false,
				ButtonMode:        true,
				AxisLeftX:         10.5,
				AxisLeftY:         0,
				AxisRightX:        0,
				AxisRightY:        0,
				AxisLeftTrigger:   0,
				AxisRightTrigger:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := &GamepadBitfield{}
			err := gb.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *gb != *tt.wantField {
				t.Errorf("UnmarshalBinary() got = %v, want %v", gb, tt.wantField)
			}
		})
	}
}