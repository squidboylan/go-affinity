package utils

import (
	"net"
)

func RandomPort() int {
	tcpPort, _ := net.Listen("tcp", ":0")
	portNum := tcpPort.Addr().(*net.TCPAddr).Port

	tcpPort.Close()
	return portNum
}
