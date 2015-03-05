package main

import (
	"com"
	"com/tcp"
	"strconv"
	"time"
)

var (
	send_ch    chan tcp.IDable
	receive_ch chan interface{}
	status_ch  chan tcp.ClientStatus
)

func main() {
	send_ch = make(chan tcp.IDable)
	receive_ch = make(chan interface{})
	status_ch = make(chan tcp.ClientStatus)
	prc := com.NewHeaderProtocol{1000}
	message := com.ElevData{1, 1, 4, 2, "up"}
	isMaster, err := com.Init(send_ch, receive_ch)
	println("er master:" + strconv.FormatBool(isMaster))
	for {
		println("data sendt")
		send_ch <- message
		time.Sleep(1 * time.Second)
	}
}

func receive() {
	message := <-receive_ch
	switch data2 := message.(type) {
	case com.ElevData:
		println("data mottatt")
	default:
		println("default")
	}
}
