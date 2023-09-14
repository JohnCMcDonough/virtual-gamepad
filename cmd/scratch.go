package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func main_2() {
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
}
