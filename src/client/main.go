package client

import (
	"fmt"
	"gamepad-agent/manager"
	"github.com/0xcafed00d/joystick"
	"log"
	"net"
	"time"
)

type GRPCAgent struct {
	conn *net.UDPConn
}

func (a GRPCAgent) init() {
	go a.forwardAction()
}

func (a GRPCAgent) forwardAction() {
	js := manager.GetJoystickObject()
	actionChannel := make(chan joystick.State)
	go manager.ReadJoystick(js, actionChannel)
	for true {
		state := <-actionChannel
		action := manager.JoystickState2Action(state)
		data := manager.Action2SwitchProtocol(action)
		_, err := a.conn.Write(data)
		if err != nil {
			log.Printf("转发时发生了错误: %v", err)
		}
	}
}

func Run() {
	conn, err := net.DialUDP(
		"udp",
		&net.UDPAddr{},
		&net.UDPAddr{
			IP:   net.ParseIP(manager.Config.Client.ServerHost),
			Port: manager.Config.Client.ServerPort,
		},
	)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	agent := GRPCAgent{conn: conn}
	agent.init()

	for {
		time.Sleep(time.Hour)
	}
}
