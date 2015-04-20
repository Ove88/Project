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
	reCalc_ch        chan com.Order
	clients          []*Client
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
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan com.Status, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 1)
	transaction_ch = make(chan interface{}, 10)
	ack_ch = make(chan com.Header, 50)
	order_ch = make(chan com.Header, 50)
	elevUpdate_ch = make(chan com.Header, 100)
	reCalc_ch = make(chan com.Order, 100)
	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch, maxNumberOfClients)
	println(localID)
	clients = append(
		clients, &Client{localID, false, false, 0, 0, make([]*com.Order, 0, maxOrderSize)})
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevStatus_ch)

	go netwMessageHandler()
	go messageSelector()
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
func messageSelector() {
	for {
		select {
		case message := <-order_ch:
			println("netwMess")
			transactionManager(&message)

		case order := <-lOrderReceive_ch:
			println("lOrder")
			order.OriginID = clients[0].ID
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, order})
		case order := <-reCalc_ch:
			println("recalc")
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, order})
		}
	}
}

func transactionManager(message *com.Header) bool {

	switch message.Data.(type) {

	case elevator.Order: // Local order
		if !clients[0].IsMaster {
			if clients[0].Active {
				message.RecvID = masterID
				send_ch <- message
			} else {
				client := calc(&message.Data)
			}
		} else {
			client := calc(&message.Data)
			return orderUpdater(&message.Data, &client, true)
		}
	case com.Order: // Remote order from client
		client := calc(&message.Data)
		return orderUpdater(&message.Data, &client, true)

	case com.ElevUpdate: // Status update from client
		for i, client := range clients {
			if message.SendID == client.ID {
				client[i].LastPosition = message.LastPosition
				client[i].Direction = message.Direction
				if message.LastPosition == client[i].Orders[0].Floor && message.Direction == -1 { // Elevator has reached its destination
					clients[i].Orders = clients[i].Orders[1:]
					orderUpdater(client[i].Orders[0], clients[i], false)
				}
				break
			}
		}

	case com.Orders: // Order update from master
		clientExists := false
		if !clients[0].IsMaster {
			for i, client := range clients {
				if client.ID == message.Data.ClientID {
					clients[i].Orders = message.Data
					clientExists = true
					if i == 0 { // Order for this elevator
						lOrderSend_ch <- clients[i].Orders[0]

						send_ch <- com.Header{
							message.MessageID, clients[0].ID, message.SendID, nil} // Ack to master
					}
					break
				}
			}
			if !clientExists {
				clients = append(clients, Client{
					message.Data.ClientID, false, false, 0, 0, message.Data})
				// send_ch <- com.Header{
				// 	message.MessageID, clients[0].ID, message.SendID, nil} // Ack to master
			}
		}
	}
}

func orderUpdater(order *com.Order, client *Client, isNewOrder bool) bool {
	orderUpdate := com.Header{
		newMessageID(), clients[0].ID, 0, com.Orders{client.ID, client.Orders}}
	for n, client_ := range clients {
		if n == 0 {
			continue
		}
		orderUpdate.RecvID = client_.ID
		send_ch <- orderUpdate

		select {
		case ack := <-ack_ch:
			if ack.MessageID != orderUpdate.MessageID || ack.SendID != client.ID {
				return false
			}
		case time.After(1 * time.Millisecond):
			return false
		}
	}
	buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{
		order.Direction, order.Floor, isNewOrder}}
	for n, client := range clients {
		if n == 0 {
			continue
		}
		buttonLightUpdate.RecvID = client.ID
		send_ch <- buttonLightUpdate
	}
	return true
}

func elevPositionManager() {
	for {
		position := <-elevPos_ch
		clients[0].LastPosition = position.LastPosition
		clients[0].Direction = position.Direction
		if !clients[0].IsMaster && clients[0].Active { // Updates master with latest position and direction
			send_ch <- com.Header{newMessageID(), clients[0].ID, masterID, position}
		}
		if position.Direction < 0 { // Elevator has stopped
			if position.LastPosition < 0 { // Stop button has been pressed
				// 	for i, order := range clients[0].Orders {
				// 		if !order.Internal {
				// 			reCalc_ch <- client.Orders[0]
				// 		}
				// 	}
			} else {

			}
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

				if status.IsMaster { // Sets master ID
					masterID = status.ID
				} else if status.ID == masterID && !status.Active { // Removes master ID
					masterID = -1
				}

				if !clients[0].Active { // Clear button lamps if elevator goes inactive
					for n, client := range clients {
						for i, order := range client.Orders {
							if n == 0 && i == 0 {
								continue
							}
							elevator.SetButtonLamp(order.Direction, order.Floor, false)
						}
					}
				}

				if clients[0].IsMaster && status.Active { // Update existing client with order lists
					message := com.Header{newMessageID(), clients[0].ID, status.ID, com.Orders{0, nil}}
					buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{0, 0, true}}

					for _, client := range clients {
						message.Data.ClientID = client.ID
						message.Data.Orders = client.Orders
						send_ch <- message
						for _, order := range client.Orders {
							if !order.Internal {
								buttonLightUpdate.Data.Button = order.Direction
								buttonLightUpdate.Data.Floor = order.Floor
								send_ch <- buttonLightUpdate
							}
						}
					}
				}

				if !status.Active && status.ID != clients[0].ID && clients[0].IsMaster { // Recalculate orders for inactive client
					for i, order := range clients[n].Orders {
						if !order.Internal {
							reCalc_ch <- client.Orders[i]
						}
					}
				}
				break
			}
		}
		if !clientExists { // New client
			clients = append(clients, Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*Order, 0, maxOrderSize)})

			if clients[0].IsMaster { // Update new client with order lists
				message := com.Header{newMessageID(), clients[0].ID, status.ID, com.Orders{0, nil}}
				buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{0, 0, true}}

				for _, client := range clients {
					message.Orders.ClientID = client.ID
					message.Data.Orders = client.Orders
					send_ch <- message
					for _, order := range client.Orders {
						if !order.Internal {
							buttonLightUpdate.Data.Button = order.Direction
							buttonLightUpdate.Data.Floor = order.Floor
							send_ch <- buttonLightUpdate
						}
					}
				}
			}
		}
	}
}

func newMessageID() int {
	return 1
}

func calc(newOrder *com.Order) Client {
	var bestClient Client
	var clientCost int
	bestCost := 1000

	for _, client := range clients {
		if ((client.LastPosition < 0) && (client.Direction < 0)) || !client.Active {
			continue
		} else {
			for _, order := range client.Orders {
				if newOrder.Direction == client.Direction {

				}
			}
		}
	}
}
