package main

import (
	"com"
	"com/tcp"
	"elevator"
	"math"
)

const (
	maxOrderSize       int     = 50
	maxNumberOfClients int     = 10
	nFloors            int     = 4
	elevConst          float32 = nFloors * (1 / 4)
	elevEstimate       float32 = 1 / 2
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
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevPos_ch)

	go netwMessageHandler()
	go channelSelector()
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
		case nil:
			println("Ack")
			ack_ch <- message
		default:
			println("Default")
			order_ch <- message
		}
	}
}
func channelSelector() {
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
		case pos := <-elevPos_ch: // Elevator position has changed
			println("elevPos")
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, pos})
		}
	}
}

func transactionManager(message *com.Header) bool {

	switch message.Data.(type) {

	case elevator.Order: // Local order
		if clients[0].Active {
			if clients[0].IsMaster {
				client := calc(&message.Data)
				return orderUpdater(&message.Data, &client, true)
			} else {
				message.RecvID = masterID
				send_ch <- message
			}
		}
	case elevator.Position: // Elevator has changed position
		clients[0].LastPosition = message.Data.LastPosition
		clients[0].Direction = message.Data.Direction

		if clients[0].Active {

			if clients[0].IsMaster {

				if message.Data.LastPosition == clients[0].Orders[0].Floor &&
					message.Data.Direction == -1 { // Elevator has reached its destination
					clients[0].Orders = clients[0].Orders[1:]
					lOrderSend_ch <- clients[0].Orders[0]
				}
			} else {
				message.RecvID = masterID
				send_ch <- message
			}
		} else if message.Data.LastPosition == clients[0].Orders[0].Floor &&
			message.Data.Direction == -1 { // Elevator has reached its destination
			// Betjene interne ordrer?
		}
	case com.Order: // Remote order from client
		client := calc(&message.Data)
		return orderUpdater(&message.Data, &client, true)

	case com.ElevUpdate: // Status update from client
		for i, client := range clients {

			if message.SendID == client.ID {
				client[i].LastPosition = message.LastPosition
				client[i].Direction = message.Direction

				if message.LastPosition == client[i].Orders[0].Floor &&
					message.Direction == -1 { // Elevator has reached its destination
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
						lOrderSend_ch <- clients[i].Orders[0] // Update elevator with order, even if no change

						send_ch <- com.Header{
							message.MessageID, clients[0].ID, message.SendID, nil} // Ack to master
					}
					break
				}
			}
			if !clientExists { //Create new client
				clients = append(clients, Client{
					message.Data.ClientID, false, false, 0, 0, message.Data})
			}
		}
	}
}

func orderUpdater(order *com.Order, client *Client, isNewOrder bool) bool {
	orderUpdate := com.Header{
		newMessageID(), clients[0].ID, 0, com.Orders{client.ID, client.Orders}}
	for n, client_ := range clients { // Update all clients with new order list
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
	for n, client_ := range clients { // Set button light for all clients
		if client_.ID != client.ID && order.Internal {
			continue
		}
		buttonLightUpdate.RecvID = client.ID
		send_ch <- buttonLightUpdate
	}
	return true
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

				if n == 0 && !status.Active { // Clear button lamps if elevator goes inactive
					for n, client := range clients {
						for i, order := range client.Orders {
							if n == 0 && i == 0 {
								continue
							}
							elevator.SetButtonLamp(order.Direction, order.Floor, false)
						}
					}
				}

				if clients[0].IsMaster && clients[0].Active && n != 0 {

					if status.Active { // Update existing client with order lists
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
					} else { // Recalculate orders for inactive client
						for i, order := range clients[n].Orders {
							if !order.Internal {
								reCalc_ch <- client.Orders[i]
							}
						}
					}
				}
				break
			}
		}
		if !clientExists { // New client connected
			clients = append(clients, Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*Order, 0, maxOrderSize)})

			if clients[0].IsMaster {
				message := com.Header{newMessageID(), clients[0].ID, status.ID, com.Orders{0, nil}}
				buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{0, 0, true}}

				for _, client := range clients { // Update new client with order lists
					message.Orders.ClientID = client.ID
					message.Data.Orders = client.Orders
					send_ch <- message
					for _, order := range client.Orders { // Set correct button lights
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
	var cost Cost

	//var bestClient Client
	var clientCost float32
	bestCost := 1000.0

	for _, client := range clients {
		if ((client.LastPosition < 0) && (client.Direction < 0)) || !client.Active {
			continue
		} else if len(client.Orders) > 0 {
			for n, order := range client.Orders {

				if newOrder.Direction == order.Direction {
					if client.LastPosition < newOrder.Floor < order.Floor { // Both orders is above elevator
						clientCost = math.Abs(newOrder.Floor - client.LastPosition)

					} else if newOrder.Floor <= client.LastPosition {
						clientCost = order.Floor - client.LastPosition + (nFloors-order.Floor)*elevEstimate
					} else {
						clientCost = newOrder.Floor - client.LastPosition
					}
				} else if n < len(client.Orders)-1 {
					continue
				} else {

				}

			}
		} else {
			clientCost = math.Abs(newOrder.Floor - client.LastPosition)
			if clientCost < bestCost {
				bestCost = clientCost
				cost = Cost{&client, clientCost, 0}
			}
		}
	}
}

type Cost struct {
	Client   *Client
	Cost     int
	OrderPos int
}
