package server

import (
	"fmt"
	"gamepad-agent/manager"
	"io/ioutil"
	"log"
	"net"
)

func Run() {
	listener, err := net.ListenUDP(
		"udp",
		&net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: manager.Config.Server.Listen,
		},
	)
	if err != nil {
		log.Fatalf("启动服务端失败: %v", err)
	}
	data := make([]byte, 8)
	for {
		n, _, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("UDP数据读取失败: %v", err)
		}
		err = ioutil.WriteFile("/dev/hidg0", data[:n], 0666)
		if err != nil {
			log.Printf("手柄操作写入失败: %v", err)
		}
	}
}
