//go:build linux
// +build linux

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/uevent"
)

func main() {
	// Create signals channel to run server until interrupted
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	decoding := make(chan bool, 1)

	sock, err := uevent.NewSocket()
	if err != nil {
		log.Fatalf("Unable to open udev socket %v", err)
	}
	defer sock.Close()

	go func() {
		log.Default().Println("Creating a new decoder!")
		dec := uevent.NewDecoder(sock)

		decoding <- true
		for {
			evt, err := dec.Decode()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(evt)
		}
	}()

	<-decoding

	enc := uevent.NewEncoder(sock)

	evt := uevent.NewUEvent("add", "/class/input/input9/mouse2", nil)
	b := enc.Encode(evt)

	_, err = sock.Write(b)

	if err != nil {
		log.Fatalf("Unable to write event to socket %v", err)
	}
	log.Printf("Wrote entire message!")

	<-done

}
