package cntr

import (
	"com"
	"elevator"
	"tcp"
)

const (
	maxOrderSize       int = 50
	maxNumberOfClients int = 10
)

var (
	send_ch          chan com.Header
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
	masterID         int
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
	send_ch = make(chan com.Header, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan com.Status, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 1)
	transaction_ch = make(chan interface{}, 10)
	ack_ch = make(chan com.Header, 50)
	order_ch = make(chan com.Header, 50)
	wait := make(chan bool)

	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch, maxNumberOfClients)
	println(localID)
	clients = append(
		clients, &Client{localID, false, false, 0, 0, make([]*com.Order, 0, maxOrderSize)})
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevStatus_ch)

	go messageHandler()
	go transactionManager()
	go orderManager()
	go elevPositionManager()
	clientStatusManager()
	//<-wait
}

func messageHandler() {
	for {
		message := <-receive_ch

		switch message.Data.(type) {
		case com.ElevUpdate:
			println("elevData")
			elevUpdate_ch <- message
		case com.Order:
			println("rOrder")
			order_ch <- message
		case com.Orders:
			println("Orders")
			orders_ch <- message
		default:
			println("default")
			ack_ch <- message
		}
	}
}
func transactionManager() {
	for {
		select {
		case order := <-order_ch:
			orderManager(order)

		case update := <-elevUpdate_ch:
			for i, client := range clients {
				if update.SendID == client.ID {
					client[i].LastPosition = update.LastPosition
					client[i].Direction = update.Direction
					break
				}
			}

		case order := <-lOrderReceive_ch:
			println("lOrder")
			orderManager(com.Header{
				newMessID(), clients[0].ID, clients[0].ID, order})
		}
	}
}

func orderManager(order com.Header) {

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

func elevPositionManager() {
	for {
		position := <-elevPos_ch
		clients[0].LastPosition = position.LastPosition
		clients[0].Direction = position.Direction
		if !clients[0].IsMaster && clients[0].Active {
			send_ch <- com.Header{newMessageID(), clients[0].ID, masterID, position}
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
				if status.IsMaster {
					masterID = status.ID
				}
				break
			}
		}
		if !clientExists {
			clients = append(clients, Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*Order, 0, maxOrderSize)})
		}
	}
}

func newMessageID() int {

}
