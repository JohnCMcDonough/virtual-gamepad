//go:build linux
// +build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JohnCMcDonough/virtual-gamepad/pkg/input"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
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

	// Create the new MQTT Server.
	server := mqtt.New(nil)

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Create a WebSocket listener on a standard port.
	ws := listeners.NewWebsocket("ws1", ":8080", nil)
	if err := server.AddListener(ws); err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP("tcp1", ":1883", nil)
	if err := server.AddListener(tcp); err != nil {
		log.Fatal(err)
	}

	gamepadHub := input.NewGamepadHub()
	backgroundCtx, cancelFn := context.WithCancel(context.Background())
	gamepadWatcher := input.NewGamepadWatcher(backgroundCtx, gamepadHub)
	gamepadWatcher.Watch(backgroundCtx)

	if err := server.AddHook(gamepadHub, nil); err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	cancelFn()
	server.Close()
}
