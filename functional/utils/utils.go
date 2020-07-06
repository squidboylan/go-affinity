package utils

import (
	"net"
)

func RandomTCPPort() int {
	tcpPort, _ := net.Listen("tcp", ":0")
	portNum := tcpPort.Addr().(*net.TCPAddr).Port

	tcpPort.Close()
	return portNum
}

func RandomUDPPort() int {
	udpAddr, _ := net.ResolveUDPAddr("udp", ":0")
	udpConn, _ := net.ListenUDP("udp", udpAddr)
	portNum := udpConn.LocalAddr().(*net.UDPAddr).Port

	udpConn.Close()
	return portNum
}
