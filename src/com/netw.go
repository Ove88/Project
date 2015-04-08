package com

import (
	"com/tcp"
	"com/udp"
	"strconv"
	"strings"
	"time"
)

const (
	localUdpPort         int = 15000
	broadcastUdpPort     int = 16000
	tcpReceiveBufferSize int = 1024
	maxNumberOfClients   int = 10
)

var (
	isMaster           bool
	localID            int
	localIP            string
	stopAnnounceMaster bool
	stopDrainUdp       bool
	udpSend_ch         chan udp.UdpPacket
	udpReceive_ch      chan udp.UdpPacket
	tcpSend_ch         <-chan tcp.IDable
	tcpReceive_ch      chan<- interface{}
	clientStatus_ch    chan tcp.ClientStatus
	config_ch          chan config
)

type config struct {
	remoteAddr string
	isMaster   bool
}
type ClientStatus struct {
	ID       int
	Active   bool
	IsMaster bool
}

func Init(send_ch <-chan tcp.IDable, receive_ch chan<- interface{},
	status_ch chan ClientStatus) (localId int, err error) {

	udpSend_ch = make(chan udp.UdpPacket, 1)
	udpReceive_ch = make(chan udp.UdpPacket, 1)
	clientStatus_ch = make(chan tcp.ClientStatus, 1)
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
	go cStatusHandler(status_ch)
	go startNetwConfig(status_ch)
	return localID, err
}

func startNetwConfig(status_ch chan ClientStatus) {

	newpr := NewHeaderProtocol{tcpReceiveBufferSize}
	go configMaster()

	configData := <-config_ch
	isMaster = configData.isMaster
	status_ch <- ClientStatus{localID, true, isMaster}

	if isMaster {
		remoteTcpPort, _ := tcp.StartServer(
			localIP, tcpSend_ch, tcpReceive_ch, clientStatus_ch, newpr, maxNumberOfClients)
		go announceMaster(remoteTcpPort)
	} else {
		tcp.StartClient(
			localIP, configData.remoteAddr, tcpSend_ch, tcpReceive_ch, clientStatus_ch, newpr)
		go drainUdpChan()
	}
}

func configMaster() {
	smallestRemoteId := 255
	stopconfig := false
	clientFound := false
	stopTimer := time.NewTimer(1 * time.Second)

	for !stopconfig {
		select {

		case <-time.After(300 * time.Millisecond):
			udpSend_ch <- udp.UdpPacket{"broadcast", []byte("ready")}

		case packet := <-udpReceive_ch:
			switch strings.Split(string(packet.Data), ":")[0] {
			case "ready":
				remoteIP := strings.Split(packet.RemoteAddr, ":")[0]
				remoteId, _ := strconv.Atoi(strings.Split(remoteIP, ".")[3])

				if remoteId != localID && !clientFound {
					clientFound = true
				}
				if remoteId < smallestRemoteId {
					smallestRemoteId = remoteId
				}
			case "connect":
				println("connect")
				remoteTcpPort := strings.Split(string(packet.Data), ":")[1]
				remoteIPAddr := strings.Split(packet.RemoteAddr, ":")[0]

				config_ch <- config{remoteIPAddr + ":" + remoteTcpPort, false}
				stopconfig = true
			}

		case <-stopTimer.C:
			if !clientFound {
				stopTimer = time.NewTimer(1 * time.Second)
			} else if localID <= smallestRemoteId {
				config_ch <- config{"", true}
				stopconfig = true
			}
		}
	}
}

func cStatusHandler(status_ch chan ClientStatus) {
	for {
		cStatus := <-clientStatus_ch
		println(cStatus.String())

		if !isMaster && cStatus.Active == false {
			status_ch <- ClientStatus{cStatus.ID, cStatus.Active, false}
			stopDrainUdp = true
			go startNetwConfig(status_ch)

		} else if cStatus.ID == localID {
			stopAnnounceMaster = true
			status_ch <- ClientStatus{localID, cStatus.Active, false}
			go startNetwConfig(status_ch)
		} else {
			status_ch <- ClientStatus{cStatus.ID, cStatus.Active, false}
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

func drainUdpChan() {
	stopDrainUdp = false
	for !stopDrainUdp {
		<-udpReceive_ch
	}
}
