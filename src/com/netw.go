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
	maxNumberOfClients int
)

type config struct {
	remoteAddr string
	isMaster   bool
}

func Init(send_ch <-chan tcp.IDable, receive_ch chan<- interface{},
	
	status_ch chan tcp.ClientStatus, maxNClients int) (int, bool) {
	var err error
	udpSend_ch = make(chan udp.UdpPacket, 1)
	udpReceive_ch = make(chan udp.UdpPacket, 1)
	clientStatus_ch = make(chan tcp.ClientStatus, 1)
	config_ch = make(chan config)
	tcpSend_ch = send_ch
	tcpReceive_ch = receive_ch
	maxNumberOfClients = maxNClients

	localIP, err = udp.Init(broadcastUdpPort, localUdpPort, udpReceive_ch, udpSend_ch)
	localID, err = strconv.Atoi(strings.Split(localIP, ".")[3])

	if err != nil {
		return 0, false
	}
	go clientStatusHandler(status_ch)
	go startNetwConfig(status_ch)
	return localID, true
}

func startNetwConfig(status_ch chan tcp.ClientStatus) {
	ok := false
	for !ok { 
		newpr := NewHeaderProtocol{tcpReceiveBufferSize}
		go configMaster()

		configData := <-config_ch
		isMaster = configData.isMaster
		status_ch <- tcp.ClientStatus{localID, true, isMaster}

		if isMaster {
			remoteTcpPort, err := tcp.StartServer(
				localIP, tcpSend_ch, tcpReceive_ch, clientStatus_ch, newpr, maxNumberOfClients)
			if err != nil{
				go announceMaster(remoteTcpPort)
				ok = true
			}
		} else {
			err := tcp.StartClient(
				localIP, configData.remoteAddr, tcpSend_ch, tcpReceive_ch, clientStatus_ch, newpr)
			if err != nil{
				go drainUdpChan()
				ok = true
			}
		}
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
			println("connect from master")
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

func clientStatusHandler(status_ch chan tcp.ClientStatus) {
	for {
		cStatus := <-clientStatus_ch
		println(cStatus.String())

		if !isMaster && cStatus.Active == false { // This is slave and master goes inactive
			status_ch <- cStatus
			if !stopDrainUdp {
				go startNetwConfig(status_ch)
			}
			stopDrainUdp = true

		} else if cStatus.ID == -1 { // Master goes inactive 
			stopAnnounceMaster = true
			cStatus.ID = localID
			status_ch <- cStatus
			go startNetwConfig(status_ch)
		} else {
			status_ch <- cStatus
		}
	}
}

//func pollMaster(masterIP string){
//	for !stopDrainUdp {
//		udpSend_ch <- udp.UdpPacket{masterIP, []byte(strconv.Itoa(localID))}
//		time.Sleep(400 * time.Millisecond)
//	}
//}

func announceMaster(masterPort int) {
	stopAnnounceMaster = false
	for !stopAnnounceMaster {
		udpSend_ch <- udp.UdpPacket{
			"broadcast", []byte("connect:" + strconv.Itoa(masterPort))}
		time.Sleep(400 * time.Millisecond)
	}
}

//func listenForClientPolls(){
//	for !stopAnnounceMaster{
//		select{
//			case <-udpReceive_ch:
//				continue
//			case <-time.After(time.Second): // Lost connection to master
//				clientStatus_ch<- tcp.ClientStatus{localID,false,false}
//		}
//	}
//}

func drainUdpChan() {
	stopDrainUdp = false
	for !stopDrainUdp {
		<-udpReceive_ch
		//select {
		//case <-udpReceive_ch:
		//	continue
		//case <-time.After(time.Second): // Lost connection to master
		//	println("lost master")
		//	clientStatus_ch <- tcp.ClientStatus{localID, false, false}
		//}
	}
}
