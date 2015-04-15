package cntr

import (
	"com"
	"elevator"
	"tcp"
)

const (
	orderSize          int = 50
	maxNumberOfClients int = 10
)

var (
	send_ch          chan tcp.IDable
	receive_ch       chan interface{}
	clientStatus_ch  chan tcp.ClientStatus
	lOrderReceive_ch chan elevator.Order
	lOrderSend_ch    chan elevator.Order
	elevPos_ch       chan elevator.Position
	order_ch         chan com.Order
	transaction_ch   chan interface{}
	ack_ch           chan com.Ack
	clients          []*Client
	allOrders        []*com.Order
	localOrders      []*com.Order
)

type Client struct {
	ID           int
	Active       bool
	IsMaster     bool
	LastPosition int
	Direction    int
	Orders       []*com.Order
}

func main() {

	clients = make([]*Client, 0, maxNumberOfClients)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan com.Status, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 1)
	transaction_ch = make(chan interface{}, 10)
	ack_ch = make(chan com.Ack, 1)
	wait := make(chan bool)

	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch)
	clients[0] = Client{localID, true, false, 0, 0, make([]*com.Order, 0, orderSize)}
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevStatus_ch)

	go elevStatusManager()
}

func transactionManager() {
	for {
		trans := <-transaction_ch

	}
}

func orderManager(order_ch chan com.Order) {
	for {
		order := <-order_ch
		if !clients[0].IsMaster {
			send_ch <- order

		} else {
			//calculate(&ordOrders []*com.Orders)
			send_ch <- order
			for client := range clients {
				//send_ch<-
			}

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
				for i, client := range clients {
					if data.SendID == client.ID {
						client[i].LastPosition = data.LastPosition
						client[i].Direction = data.Direction
					}
				}
			case com.Order:
				println("rOrder")
				order_ch <- data
			case com.Ack:
				println("Ack")
				ack_ch <- data
			default:
				println("default")
			}

		case order := <-lOrderReceive_ch:
			println("lOrder")
			order_ch <- com.Order{
				newMessID(), clients[0].ID, clients[0].ID, order.Internal, order.Floor, order.Direction}
		}
	}
}

func elevPositionManager() {
	for {
		position := <-elevPos_ch
		clients[0].LastPosition = position.LastPos
		clients[0].Direction = position.Direction
		if !clients[0].IsMaster && clients[0].Active {
			send_ch <- com.ElevUpdate{newMessageID()}
		}
	}
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
				clients[n].IsMaster = status.IsMaster
				break
			}
		}

		if !clientExists {
			client := Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*Order, 0, orderSize)}
			clients = append(clients, &client)
		}
	}
}

func newMessageID() int {

}
