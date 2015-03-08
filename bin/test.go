package main

import (
	"com"
	"com/tcp"
	"com/udp"
	"strconv"
	"time"
)

var status_ch chan tcp.ClientStatus

func main() {
	send_ch := make(chan tcp.IDable)
	receive_ch := make(chan interface{})
	status_ch = make(chan tcp.ClientStatus)
	prc := com.NewHeaderProtocol{1000}
	message := com.ElevData{1, 1, 4, 2, "up"}
	bjarne()
	go listenStatus()
	//tcp.StartServer("127.0.0.1", 12000, send_ch, receive_ch, prc, 5)
	tcp.StartClient("127.0.0.1", "127.0.0.1:12000", 12001, send_ch, receive_ch, status_ch, prc)
	println("startet")
	go func(message tcp.IDable) {
		for {
			send_ch <- message
			time.Sleep(1 * time.Second)
		}
	}(message)
	for {
		data := <-receive_ch
		switch data2 := data.(type) {
		case com.ElevData:
			println("data:" + data2.Direction)

		default:
			println("default")
		}
	}
}
func listenStatus() {
	for {
		status := <-status_ch
		println("status: " + strconv.Itoa(status.ID) + " " + strconv.FormatBool(status.Active))
	}
}

func bjarne() {
	udpSend_ch := make(chan udp.UdpPacket, 5)
	udpReceive_ch := make(chan udp.UdpPacket, 5)
	udp.Init(10000, 11000, udpReceive_ch, udpSend_ch)
	for {
		udpSend_ch <- udp.UdpPacket{"127.0.0.1:11000", []byte("test2")}
		pack := <-udpReceive_ch
		println(string(pack.Data))
		time.Sleep(1 * time.Second)

	}
	time.Sleep(1 * time.Second)
}
