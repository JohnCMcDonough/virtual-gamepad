//go:build linux
// +build linux

package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	input "github.com/JohnCMcDonough/virtual-gamepad/pkg/input"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

//go:embed build
var httpStaticContent embed.FS

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
	ws := listeners.NewWebsocket("ws1", ":8081", nil)
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

	// load static content embedded in the app
	contentStatic, _ := fs.Sub(httpStaticContent, "build")
	http.Handle("/", http.FileServer(http.FS(contentStatic)))

	go func() {
		log.Print("Listening on :8080...")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := os.Stat("/run/udev/control"); os.IsNotExist(err) {
		server.Log.Warn().Msg("The /run/udev/control file does not exist. Currently running applications may not receive events correctly. Creating...")
		if err = os.MkdirAll("/run/udev", 0o666); err != nil {
			server.Log.Err(err).Msg("Failed to create directory /run/udev")
		}
		if _, err = os.Create("/run/udev/control"); err != nil {
			server.Log.Err(err).Msg("Failed to create file /run/udev/control")
		} else {
			server.Log.Info().Msg("Created /run/udev/control successfully!")
		}
	}

	<-done
	cancelFn()
	server.Close()
}
