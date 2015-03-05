package main

import (
	"com"
	//"com/udp"
	"com/tcp"
	"strconv"
	"time"
)

var status_ch chan tcp.ClientStatus

func main() {
	send_ch := make(chan tcp.IDable, 5)
	receive_ch := make(chan interface{}, 5)
	status_ch = make(chan tcp.ClientStatus)
	prc := com.NewHeaderProtocol{1000}
	//message := com.ElevData{1, 102, 4, 2, "up"}
	go listenStatus()
	tcp.StartServer("127.0.0.1", 12000, send_ch, receive_ch, status_ch, prc, 5)
	//tcp.StartClient("127.0.0.1", "127.0.0.1:12000", 12001, send_ch, receive_ch, prc)
	println("startet")

	for {
		data := <-receive_ch
		println("venter")
		switch data2 := data.(type) {
		case com.ElevData:
			println("data mottatt")
			send_ch <- data2

		default:
			println("default")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
func listenStatus() {
	for {
		status := <-status_ch
		println("status: " + strconv.Itoa(status.ID) + " " + strconv.FormatBool(status.Active))
	}
}
