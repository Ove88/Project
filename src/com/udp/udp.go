package udp

import (
	"net"
	"strings"
	"strconv"
)

const maxPacketSize int = 100

var (
	baddr *net.UDPAddr
	laddr *net.UDPAddr
)

type UdpPacket struct {
	RemoteAddr string
	Data       []byte
}

func Init(broadcastPort, localPort int, receive_ch chan<- UdpPacket,
	send_ch <-chan UdpPacket) (string, error) {

	taddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastPort))
	tempConn, err := net.DialUDP("udp4", nil, taddr)
	defer tempConn.Close()

	tempAddr := tempConn.LocalAddr()
	//laddr, err = net.ResolveUDPAddr("udp4", "127.0.0.1:"+strconv.Itoa(localPort))
	laddr, err = net.ResolveUDPAddr("udp4", tempAddr.String())
	laddr.Port = localPort

	bTemp := strings.SplitAfterN(laddr.IP.String(), ".", 4)
	broadcastIP := bTemp[0] + bTemp[1] + bTemp[2] + "255"
	baddr, err = net.ResolveUDPAddr("udp4", broadcastIP+":"+strconv.Itoa(broadcastPort))

	bConn, err := net.ListenUDP("udp4", baddr)
	lConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		bConn.Close()
		lConn.Close()
		return "", err
	}
	go receivePackets(bConn, receive_ch)
	go receivePackets(lConn, receive_ch)
	go sendPackets(lConn, send_ch)

	return laddr.IP.String(), err
}

func receivePackets(conn *net.UDPConn, receive_ch chan<- UdpPacket) {

	buffer := make([]byte, maxPacketSize)
	for {
		n, raddr, _ := conn.ReadFromUDP(buffer)
		receive_ch <- UdpPacket{raddr.String(), buffer[:n]}
	}
}

func sendPackets(conn *net.UDPConn, send_ch <-chan UdpPacket) {

	for {
		packet := <-send_ch
		if packet.RemoteAddr == "broadcast" {
			conn.WriteToUDP(packet.Data, baddr)
		} else {
			raddr, _ := net.ResolveUDPAddr("udp4", packet.RemoteAddr)

			_, err := conn.WriteToUDP(packet.Data, raddr)
			if err != nil {
				panic(err)
			}
		}
	}
}
