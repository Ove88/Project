package cntr

import (
	"com"
	"elevator"
)

const (
	orderSize          int = 50
	maxNumberOfClients int = 10
)

var (
	send_ch         chan tcp.IDable
	receive_ch      chan interface{}
	clientStatus_ch chan com.ClientStatus
	lOrder_ch       chan elevator.Order
	elevStatus_ch   chan elevator.Status
	clients         []*Client
	isMaster        bool
)

type Client struct {
	ID        int
	Active    bool
	Position  int
	Direction int
	Orders    []*com.Order
}

func main() {

	clients = make([]*Client, 0, maxNumberOfClients)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan com.Status, 5)
	lOrder_ch = make(chan elevator.Order, 5)
	elevStatus_ch = make(chan elevator.Status, 5)
	wait := make(chan bool)
	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch)
	clients[0] = Client{localID, true, 0, 0, make([]*com.Order, 0, orderSize)}
}

func orderManager() {
	for {
		order := <-order_ch
		calculate(&order)
		for client := range clients {

		}

	}
}

func messageHandler() {
	for {
		select {
		case message := <-receive_ch:

			switch data := message.(type) {
			case com.ElevUpdate:
				println("elevData")
			case com.Order:
				println("rOrder")
			case com.Ack:
				println("Ack")
			default:
				println("default")
			}

		case order := <-lOrder_ch:
			println("lOrder")
		}
	}
}

func elevatorDataManager() {
	data := <-elevUpdate_ch
}

func clientStatusManager() {
	var clientExists bool
	for {
		clientExists = false
		status := <-clientStatus_ch
		for n, client := range clients {
			if client.ID == status.ID {
				clientExists = true
				clients[n].Active = status.Active
				if n == 0 {
					isMaster = status.IsMaster
				}
				break
			}
		}

		if !clientExists {
			client := Client{
				status.ID, status.Active, 0, 0, make([]*Order, 0, orderSize)}
			clients = append(clients, &client)
		}
	}
}
