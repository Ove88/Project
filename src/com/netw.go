package com

import (
	"com/tcp"
	"com/udp"
	"strconv"
	"strings"
	"time"
)

var (
	localID       int
	udpSend_ch    chan udp.UdpPacket
	udpReceive_ch chan udp.UdpPacket
	status_ch     chan tcp.ClientStatus
)

const (
	localTcpListenPort   int = 14000
	localUdpPort         int = 15000
	broadcastUdpPort     int = 16000
	tcpReceiveBufferSize int = 1024
	maxNumberOfClients   int = 10
)

type ElevData struct {
	Transaction_id int
	Client_id      int
	State          int
	Position       int
	Direction      string
	//Destinations   []int
}

func (e ElevData) RemoteID() int {
	return e.Client_id
}
func (e ElevData) String() string {
	return "TID:" + strconv.Itoa(e.Transaction_id) + ", ClientID:" +
		strconv.Itoa(e.Client_id) + ", State:" + e.Direction
}

/////   Sett inn flere datastructer her   /////

// type ElevOrder struct {
// 	Client_id     	 int
// 	Elev_destination int
//}

func Init(send_ch <-chan tcp.IDable, receive_ch chan<- interface{}) (isMaster bool, err error) {

	newpr := NewHeaderProtocol{tcpReceiveBufferSize}
	udpSend_ch = make(chan udp.UdpPacket, 5)
	udpReceive_ch = make(chan udp.UdpPacket, 5)
	status_ch = make(chan tcp.ClientStatus, 1)

	localIP, err := udp.Init(broadcastUdpPort, localUdpPort, udpReceive_ch, udpSend_ch)
	localID, err = strconv.Atoi(strings.Split(localIP, ".")[3])
	println(localID)
	if err != nil {
		//return
		println(err.Error())
	}

	masterAddr, isMaster := masterConfig()

	if isMaster {
		err = tcp.StartServer(
			localIP, localTcpListenPort, send_ch, receive_ch, status_ch, newpr, maxNumberOfClients)
		go announceMaster()
		go readUDP()
	} else {
		err = tcp.StartClient(
			localIP, masterAddr, localTcpListenPort, send_ch, receive_ch, status_ch, newpr)
	}
	return
}

// func status_handler() {
// 	for {
// 		status := <-status_ch

// 	}
// }

func announceMaster() {
	for {
		udpSend_ch <- udp.UdpPacket{
			"broadcast", []byte("connect:" + strconv.Itoa(localTcpListenPort))}
		time.Sleep(500 * time.Millisecond)
	}
}
func readUDP() {
	for {
		p := <-udpReceive_ch
		println(string(p.Data))
	}
}
func masterConfig() (string, bool) {
	smallestRemoteId := 255
	stopSending := false
	stopTimer := time.NewTimer(2 * time.Second)
	for {
		select {

		case <-time.After(200 * time.Millisecond):
			println("sender")
			if !stopSending {
				udpSend_ch <- udp.UdpPacket{"broadcast", []byte("ready")}
			}
		case packet := <-udpReceive_ch:
			switch strings.Split(string(packet.Data), ":")[0] {
			case "ready":
				remoteIP := strings.Split(packet.RemoteAddr, ":")[0]
				remoteId, _ := strconv.Atoi(strings.Split(remoteIP, ".")[3])
				if remoteId < smallestRemoteId {
					smallestRemoteId = remoteId
				}
			case "connect":
				println("connect")
				remoteTcpPort := strings.Split(string(packet.Data), ":")[1]
				remoteIPAddr := strings.Split(packet.RemoteAddr, ":")[0]
				return remoteIPAddr + ":" + remoteTcpPort, false
			}

		case <-stopTimer.C:
			println("ferdig")
			stopSending = true
			if localID <= smallestRemoteId {
				return "", true
			}
		}
	}
}
