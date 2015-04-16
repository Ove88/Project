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
	send_ch          chan tcp.IDable
	receive_ch       chan interface{}
	clientStatus_ch  chan tcp.ClientStatus
	lOrderReceive_ch chan elevator.Order
	lOrderSend_ch    chan elevator.Order
	elevPos_ch       chan elevator.Position
	order_ch         chan com.Order
	transaction_ch   chan interface{}
	ack_ch           chan com.Header
	elevUpdate_ch    chan com.Header
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

	masterID = -1
	clients = make([]*Client, 0, maxNumberOfClients)
	allOrders = make([]*com.Order, 0, maxOrderSize*maxNumberOfClients)
	localOrders = make([]*com.Order, 0, maxOrderSize)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan com.Status, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 1)
	transaction_ch = make(chan interface{}, 10)
	ack_ch = make(chan com.Header, 50)
	order_ch = make(chan com.Header, 50)
	elevUpdate_ch = make(chan com.Header, 5)

	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch, maxNumberOfClients)
	println(localID)
	clients = append(
		clients, &Client{localID, false, false, 0, 0, make([]*com.Order, 0, maxOrderSize)})
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevStatus_ch)

	go netwMessageHandler()
	go messageManager()
	go elevPositionManager()
	clientStatusManager()
}

func netwMessageHandler() {
	for {
		message := <-receive_ch

		switch message.Data.(type) {
		case com.ButtonLamp:
			println("buttonLamp")
			elevator.SetButtonLamp(
				message.Data.Button, message.Data.Floor, message.Data.State)
		// case com.ElevUpdate:
		// 	println("elevData")
		// 	elevUpdate_ch <- message
		// case com.Order:
		// 	println("rOrder")
		// 	order_ch <- message
		// case com.Orders:
		// 	println("Orders")
		// 	orders_ch <- message
		case nil:
			println("Ack")
			ack_ch <- message
		default:
			println("Default")
			order_ch <- message
		}
	}
}
func messageManager() {
	for {
		select {
		case message := <-order_ch:
			println("netwMess")
			transactionManager(&message)

		case order := <-lOrderReceive_ch:
			println("lOrder")
			transactionManager(&com.Header{
				newMessID(), clients[0].ID, 0, order})
		}
	}
}

func transactionManager(message *com.Header) bool {

	switch message.Data.(type) {

	case elevator.Order: // Local Order
		if message.Data.Internal {
			if !clients[0].IsMaster && clients[0].Active {
				message.RecvID = masterID
				send_ch <- message
			}
			// gi ordre til heis
		} else {
			if !clients[0].Active {
				return false
			} else if !clients[0].IsMaster {
				message.RecvID = masterID
				send_ch <- message
				return true
			} else {
				return orderHandler(&message)
			}
		}
	case com.Order: // Remote Order
		if message.Data.Internal {
			// oppdater ordreliste for heis
		} else {
			if clients[0].IsMaster {
				return orderHandler(&message)
			} else {
				//gi ordre til heis.
			}
		}
	case com.ElevUpdate:
		for i, client := range clients {
			if update.SendID == client.ID {
				client[i].LastPosition = update.LastPosition
				client[i].Direction = update.Direction
				break
			}
		}
	case com.Orders:
		if !clients[0].IsMaster {
			allOrders = message.Data
		}
	}
}

func orderHandler(message *com.Header) bool {
	//calculate(&ordOrders []*com.Orders)
	//message.RecvID =
	//message.SendID =
	send_ch <- message
	select {
	case ack := <-ack_ch:
		if ack.MessageID != message.MessageID {
			//error
		}
	case time.After(1 * time.Millisecond):
		//error
	}
	for n, client := range clients {
		if n == 0 {
			continue
		}
		send_ch <- com.Header{newMessageID(), clients[0].ID, client.ID, allOrders}
		select {
		case ack := <-ack_ch:
			if ack.MessageID != message.MessageID {
				//error
			}
		case time.After(1 * time.Millisecond):
			//error
		}
	}
	for n, client := range clients {
		if n == 0 {
			continue
		}
		send_ch <- com.Header{newMessageID(), clients[0].ID, client.ID, com.ButtonLamp{
			message.Direction, message.Floor, true}}
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
				} else if status.ID == masterID && !status.Active {
					masterID = -1
				}
				if clients[0].IsMaster && status.Active {
					send_ch <- com.Header{newMessageID(), clients[0].ID, status.ID, allOrders}
				}
				break
			}
		}
		if !clientExists {
			clients = append(clients, Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*Order, 0, maxOrderSize)})
			if clients[0].IsMaster {
				send_ch <- com.Header{newMessageID(), clients[0].ID, status.ID, allOrders}
			}
		}
	}
}

func newMessageID() int {
	return 1
}
