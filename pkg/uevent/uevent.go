package uevent

import (
	"log"

	"github.com/pilebones/go-udev/netlink"
)

var logger = log.Default()

func PrintUEvent(evt netlink.UEvent) {
	logger.Printf("Action: %v", evt.Action)
	logger.Printf("KObj: %v", evt.KObj)
	logger.Printf("Env: %v", evt.Env)
}
