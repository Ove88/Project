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
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{})
	status_ch = make(chan tcp.ClientStatus)
	message := com.ElevData{1, 1, 4, 2, "up"}
	isMaster, err := com.Init(send_ch, receive_ch)
	if err != nil {
		println(err.Error())
	}
	println("er master:" + strconv.FormatBool(isMaster))
	if !isMaster {
		for {
			println("data sendt")
			send_ch <- message
			time.Sleep(1 * time.Second)
		}
	} else {
		go receive()
	}
}

func receive() {
	message := <-receive_ch
	for {
		switch data2 := message.(type) {
		case com.ElevData:
			println("data mottatt:" + data2.String())
		default:
			println("default")
		}
	}
}
