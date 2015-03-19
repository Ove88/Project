package cntr

import (
	"com"
	"elevator"
)

const orderSize int = 50

var (
	send_ch       chan tcp.IDable
	receive_ch    chan interface{}
	status_ch     chan com.Status
	lOrder_ch     chan elevator.Order
	elevStatus_ch chan elevator.Status
	clients       []*Client
)

type Client struct {
	ID        int
	Active    bool
	Position  int
	Direction int
	Orders    []*com.Order
}

func transactionManager() {
	for {
		order := <-order_ch

	}
}

func packetHandler() {
	for {
		for i := 0; i < count; i++ {

		}
		select {
		case message := <-receive_ch:

			switch data := message.(type) {
			case com.ElevData:
				println("elevData")
			case com.Order:
				println("rOrder")
			default:
				println("default")
			}

		case status := <-status_ch:
			println("status")

		case order := <-lOrder_ch:
			println("lOrder")
		}
	}
}

func statusHandler() {
	var clientExists bool
	for {
		clientExists = false
		status := <-status_ch
		if status.ID == localID {

		} else {
			for n, client := range clients {
				if client.ID == status.ID {
					clientExists = true
					clients[n].Active = status.Active
					break
				}
			}
		}
		if !clientExists {
			client := Client{
				status.ID, status.Active, 0, 0, make([]*Order, 0, orderSize)}
			clients = append(clients, &client)
		}
	}
}

func main() {
	for {
		select {
		case <-receive_ch:
		}
	}
}
