package main

import (
	"gamepad-agent/client"
	"gamepad-agent/manager"
	"gamepad-agent/server"
	"log"
)

func main() {
	log.Printf("type: %s", manager.Config.Type)
	switch manager.Config.Type {
	case "server":
		server.Run()
	case "client":
		client.Run()
	}
}
