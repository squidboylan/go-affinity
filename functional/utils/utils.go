package utils

import (
	"net"
)

func RandomTCPPort() (int, error) {
	tcpPort, err := net.Listen("tcp", ":0")
	if err != nil{
		return 0, err
	}
	portNum := tcpPort.Addr().(*net.TCPAddr).Port

	tcpPort.Close()
	return portNum, nil
}

func RandomUDPPort() (int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil{
		return 0, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil{
		return 0, err
	}

	portNum := udpConn.LocalAddr().(*net.UDPAddr).Port

	udpConn.Close()
	return portNum, nil
}
