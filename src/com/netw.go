package com

import (
	"com/tcp"
	"com/udp"
	"strconv"
	"strings"
	"time"
)

var (
	isMaster           bool
	localID            int
	localIP            string
	stopAnnounceMaster bool
	udpSend_ch         chan udp.UdpPacket
	udpReceive_ch      chan udp.UdpPacket
	tcpSend_ch         <-chan tcp.IDable
	tcpReceive_ch      chan<- interface{}
	cStatus_ch         chan tcp.ClientStatus
	config_ch          chan config
)

const (
	localUdpPort         int = 15000
	broadcastUdpPort     int = 16000
	tcpReceiveBufferSize int = 1024
	maxNumberOfClients   int = 10
)

type config struct {
	remoteAddr string
	isMaster   bool
}
type Status struct {
	ID     int
	Active bool
}

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

func Init(send_ch <-chan tcp.IDable, receive_ch chan<- interface{}, status_ch chan Status) (err error) {

	udpSend_ch = make(chan udp.UdpPacket, 1)
	udpReceive_ch = make(chan udp.UdpPacket, 1)
	cStatus_ch = make(chan tcp.ClientStatus, 1)
	config_ch = make(chan config)
	tcpSend_ch = send_ch
	tcpReceive_ch = receive_ch

	localIP, err = udp.Init(broadcastUdpPort, localUdpPort, udpReceive_ch, udpSend_ch)
	localID, err = strconv.Atoi(strings.Split(localIP, ".")[3])
	println(localID)
	if err != nil {
		//return
		println(err.Error())
	}
	go status_handler(status_ch)
	go startConfig(status_ch)
	return
}

func startConfig(status_ch chan Status) {

	newpr := NewHeaderProtocol{tcpReceiveBufferSize}
	go configMaster()

	configData := <-config_ch
	isMaster = configData.isMaster
	if isMaster {
		remoteTcpPort, _ := tcp.StartServer(
			localIP, tcpSend_ch, tcpReceive_ch, cStatus_ch, newpr, maxNumberOfClients)
		go announceMaster(remoteTcpPort)
	} else {
		tcp.StartClient(
			localIP, configData.remoteAddr, tcpSend_ch, tcpReceive_ch, cStatus_ch, newpr)
	}
}

func status_handler(status_ch chan Status) {
	for {
		cStatus := <-cStatus_ch
		println(cStatus.String())

		if !isMaster && cStatus.Active == false {
			go startConfig(status_ch)
		} else if cStatus.ID == -1 {
			stopAnnounceMaster = true
			go startConfig(status_ch)
		} else {
			status_ch <- Status{cStatus.ID, cStatus.Active}
		}
	}
}

func announceMaster(masterPort int) {
	stopAnnounceMaster = false
	for !stopAnnounceMaster {
		udpSend_ch <- udp.UdpPacket{
			"broadcast", []byte("connect:" + strconv.Itoa(masterPort))}
		time.Sleep(500 * time.Millisecond)
	}
}

func configMaster() {
	smallestRemoteId := 255
	stopSending := false
	startConfig := true
	var stopTimer *time.Timer
	for {
		select {

		case <-time.After(300 * time.Millisecond):
			if !stopSending {
				udpSend_ch <- udp.UdpPacket{"broadcast", []byte("ready")}
			}
		case packet := <-udpReceive_ch:
			switch strings.Split(string(packet.Data), ":")[0] {
			case "ready":
				remoteIP := strings.Split(packet.RemoteAddr, ":")[0]
				remoteId, _ := strconv.Atoi(strings.Split(remoteIP, ".")[3])

				if remoteId != localID && startConfig {
					stopTimer = time.NewTimer(1 * time.Second)
					startConfig = false
				}
				if remoteId < smallestRemoteId {
					smallestRemoteId = remoteId
				}
			case "connect":
				remoteTcpPort := strings.Split(string(packet.Data), ":")[1]
				remoteIPAddr := strings.Split(packet.RemoteAddr, ":")[0]
				config_ch <- config{remoteIPAddr + ":" + remoteTcpPort, false}
				break
			}

		case <-stopTimer.C:
			stopSending = true
			if localID <= smallestRemoteId {
				config_ch <- config{"", true}
				break
			}
		}
	}
}
