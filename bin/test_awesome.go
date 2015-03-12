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
	status_ch  chan com.Status
	active     bool
)

func main() {
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{})
	status_ch = make(chan com.Status, 1)
	message := com.ElevData{1, 1, 4, 2, "up"}
	err := com.Init(send_ch, receive_ch, status_ch)
	if err != nil {
		println(err.Error())
	}
	stat := <-status_ch
	println("er master:" + strconv.FormatBool(stat.Active))
	go status_listener()
	if !stat.Active {
		for {
			if active {
				println("data sendt")
				send_ch <- message
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		receive()

	}
}

func receive() {
	for {
		message := <-receive_ch
		switch data2 := message.(type) {
		case com.ElevData:
			println("data mottatt:" + data2.String())
		default:
			println("default")
		}
	}
}
func status_listener() {
	for {
		status := <-status_ch
		active = status.Active
	}
}
