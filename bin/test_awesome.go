package main

import (
	"com"
	"com/tcp"
	"strconv"
	"time"
)

const maxNumberOfClients int = 10

var (
	send_ch    chan tcp.IDable
	receive_ch chan interface{}
	status_ch  chan com.Status
	active     bool
	localID    int
	clients    []*Client
)

type Client struct {
	Id     int
	Active bool
}

func main() {
	clients = make([]*Client, 0, maxNumberOfClients)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{})
	status_ch = make(chan com.Status, 1)
	wait := make(chan bool)
	localID, _ = com.Init(send_ch, receive_ch, status_ch)
	// if err != nil {
	// 	println(err.Error())
	// }
	//localID = stat.ID
	go status_listener()
	go receive()
	<-wait
}

func send(client_ *Client) {
	message := com.ElevData{1, client_.Id, 4, 2, "up"}
	for {
		if client_.Active {
			println("data sendt")
			send_ch <- message
		}
		time.Sleep(1 * time.Second)
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
	var exists bool
	for {
		exists = false
		status := <-status_ch
		for n, _ := range clients {
			if status.ID == clients[n].Id {
				clients[n].Active = status.Active
				exists = true
				break
			}
			if !exists {
				println("ny")
				client_ := Client{status.ID, status.Active}
				clients = append(clients, &client_)
				go send(&client_)
			}
		}
		if status.ID == localID {
			println("er master:" + strconv.FormatBool(status.IsMaster))

		}
	}
}
