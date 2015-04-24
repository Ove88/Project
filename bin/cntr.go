package main

import (
	"com"
	"com/tcp"
	"elevator"
	"math"
	"time"
	"strconv"
)

const (
	maxOrderSize       int = 50
	maxNumberOfClients int = 10
	nFloors            int = 4
)

var (
	send_ch          chan tcp.IDable
	receive_ch       chan interface{}
	clientStatus_ch  chan tcp.ClientStatus
	lOrderReceive_ch chan elevator.Order
	lOrderSend_ch    chan elevator.Order
	elevPos_ch       chan elevator.Position
	message_ch       chan com.Header
	ack_ch           chan com.Header
	elevUpdate_ch    chan com.Header
	reCalc_ch        chan *com.Order
	clients          []*Client
	masterID         int
	isAlone          bool
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
	isAlone = true
	clients = make([]*Client, 0, maxNumberOfClients)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan tcp.ClientStatus, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 1)
	ack_ch = make(chan com.Header, 50)
	message_ch = make(chan com.Header, 50)
	elevUpdate_ch = make(chan com.Header, 100)
	reCalc_ch = make(chan *com.Order, 100)

	localID, _ := com.Init(send_ch, receive_ch, clientStatus_ch, maxNumberOfClients)
	clients = append(
		clients, &Client{localID, true, true, 0, 0, make([]*com.Order, 0, maxOrderSize)})
	elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevPos_ch)
	println(localID)

	go netwMessageHandler()
	go channelSelector()
	clientStatusManager()
}

func netwMessageHandler() {
	for {
		message_ := <-receive_ch
		message := message_.(com.Header)

		switch data := message.Data.(type) {
		case com.ButtonLamp: // Button light update from master
			println("buttonLamp")
			elevator.SetButtonLamp(
				data.Button, data.Floor, data.State)
		case nil: // Ack message from client
			println("Ack")
			ack_ch <- message
		default: // All other messages
			println("Default")
			message_ch <- message
		}
	}
}
func channelSelector() {
	for {
		select {
		case message := <-message_ch: // Messages from the network
			println("netwMess")
			transactionManager(&message)

		case order := <-lOrderReceive_ch: // Local orders from elevator
			//println("lOrder")
			order.OriginID = clients[0].ID
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, order})
		case order := <-reCalc_ch: // Recalculate orders from inactive client
			println("recalc")
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, order})
		case pos := <-elevPos_ch: // Elevator position has changed
			//println("elevPos")
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, pos})
		}
	}
}

func transactionManager(message *com.Header) bool {

	switch data := message.Data.(type) {

	case elevator.Order: // Local order
		if clients[0].Active {
			if clients[0].IsMaster {
				chosenClient := calc(elevToCom(&data))
				println(strconv.Itoa(chosenClient.ID))
				if orderUpdater(elevToCom(&data), &chosenClient, true) {
					for n, client_ := range clients {
						if client_.ID == chosenClient.ID {
							clients[n].Orders = chosenClient.Orders
						}
					}
				}
			} else {
				message.RecvID = masterID
				send_ch <- message
			}
		}
	case elevator.Position: // Elevator has changed position
		clients[0].LastPosition = data.LastPosition
		clients[0].Direction = data.Direction
		if clients[0].Active {
			if clients[0].IsMaster {
				if len(clients[0].Orders) > 0 {
					if data.LastPosition == clients[0].Orders[0].Floor &&
						data.Direction == -1 { // Elevator has reached its destination
						println("ankommet etasje")
						clients[0].Orders = clients[0].Orders[1:]
						if len(clients[0].Orders) > 0 {
							lOrderSend_ch <- comToElev(clients[0].Orders[0])
						}
					}
				}
			} else {
				message.RecvID = masterID
				send_ch <- message
			}
			//} else if data.LastPosition == clients[0].Orders[0].Floor &&
			//	data.Direction == -1 { // Elevator has reached its destination
			//	// Betjene interne ordrer?
		}
	case com.Order: // Remote order from client
		chosenClient := calc(&data)
		return orderUpdater(&data, &chosenClient, true)

	case com.ElevUpdate: // Status update from client
		for i, client := range clients {

			if message.SendID == client.ID {
				clients[i].LastPosition = data.LastPosition
				clients[i].Direction = data.Direction
				if len(clients[0].Orders) > 0 {
					if data.LastPosition == clients[i].Orders[0].Floor &&
						data.Direction == -1 { // Elevator has reached its destination
						lastOrder := clients[i].Orders[0]
						clients[i].Orders = clients[i].Orders[1:]
						orderUpdater(lastOrder, clients[i], false) //bug

					}
				}
				break
			}
		}
	case com.Orders: // Order update from master
		clientExists := false
		if !clients[0].IsMaster {
			for i, client := range clients {
				if client.ID == data.ClientID {
					clients[i].Orders = data.Orders
					clientExists = true
					if i == 0 { // Order for this elevator
						lOrderSend_ch <- comToElev(clients[i].Orders[0]) // Update elevator with order, even if no change

						send_ch <- com.Header{
							message.MessageID, clients[0].ID, message.SendID, nil} // Ack to master
					}
					return true
				}
			}
			if !clientExists { //Create new client
				clients = append(clients, &Client{
					data.ClientID, false, false, 0, 0, data.Orders})
			}
		}
	}
	return true
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
		if isNewOrder && !isAlone {
			select {
			case ack := <-ack_ch:
				if ack.MessageID != orderUpdate.MessageID || ack.SendID != client.ID {
					return false
				}
			case <-time.After(1 * time.Millisecond):
				return false
			}
		}
	}
	buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{
		order.Direction, order.Floor, isNewOrder}}
	for n, client_ := range clients { // Set button light for clients
		if (client_.ID != client.ID && order.Internal) || n == 0 {
			continue
		}
		buttonLightUpdate.RecvID = client.ID
		send_ch <- buttonLightUpdate
	}
	elevator.SetButtonLamp(order.Direction, order.Floor, isNewOrder)   // Set button light for this elevator
	if (client.ID == clients[0].ID){								    //  Update master elevator
		lOrderSend_ch <- comToElev(client.Orders[0])							
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
						message := com.Header{newMessageID(), clients[0].ID, status.ID, nil}
						buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}

						for _, client := range clients {
							message.Data = com.Orders{client.ID, client.Orders}
							send_ch <- message
							for _, order := range client.Orders {
								if !order.Internal {
									buttonLightUpdate.Data = com.ButtonLamp{order.Direction, order.Floor, true}
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
			isAlone = false
			clients = append(clients, &Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*com.Order, 0, maxOrderSize)})

			if clients[0].IsMaster {
				message := com.Header{newMessageID(), clients[0].ID, status.ID, nil}
				buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}

				for _, client := range clients { // Update new client with order lists
					message.Data = com.Orders{client.ID, client.Orders}
					send_ch <- message
					for _, order := range client.Orders { // Set correct button lights
						if !order.Internal {
							buttonLightUpdate.Data = com.ButtonLamp{order.Direction, order.Floor, true}
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

func elevToCom(order *elevator.Order) *com.Order {
	return &com.Order{order.OriginID, order.Internal, order.Floor, order.Direction}
}

func comToElev(order *com.Order) elevator.Order {
	return elevator.Order{order.OriginID, order.Internal, order.Floor, order.Direction}
}

type Cost struct {
	Client   *Client
	Cost     int
	OrderPos int
}

func calc(newOrder *com.Order) Client {
	//return Client{clients[0].ID,clients[0].Active,clients[0].IsMaster,clients[0].LastPosition,clients[0].Direction,clients[0].Orders}
	const stopCost int = 1
	var cost Cost
	var bestCost int
	var clientCost int
	var pos int
	bestCost = 1000

	for _, client := range clients {
		if ((client.LastPosition < 0) && (client.Direction < 0)) || !client.Active || (client.ID != newOrder.OriginID && newOrder.Internal){
			continue
		} else if len(client.Orders) > 0 {

			for n, order := range client.Orders {
				if newOrder.Direction == order.Direction {
					if newOrder.Floor <= order.Floor && newOrder.Floor > client.LastPosition { // case 1,7
						if order.Direction == 0 {
							pos = n
							clientCost = newOrder.Floor - client.LastPosition + (n+1)*stopCost
						} else {
							pos = n+1
							clientCost = 2*order.Floor - client.LastPosition - newOrder.Floor + (n+1)*stopCost // case 7
						}
					} else if newOrder.Floor <= order.Floor && newOrder.Floor < client.LastPosition { // case 3,5
						if order.Direction == 0 {
							pos = n
							clientCost = client.LastPosition - newOrder.Floor
						} else {
							pos = n+1
							clientCost = client.LastPosition - newOrder.Floor + (n+1)*stopCost
						}
					} else if newOrder.Floor > client.LastPosition { // case 2,8
						if order.Direction == 0 {
							pos = n+1
							clientCost = newOrder.Floor - client.LastPosition + (n+1)*stopCost
						} else {
							pos = n
							clientCost = newOrder.Floor - client.LastPosition
						}
					} else if newOrder.Floor < client.LastPosition { // case 4,6
						if order.Direction == 0 {
							pos = n+1
							clientCost = client.LastPosition - 2*order.Floor + newOrder.Floor + (n+1)*stopCost
						} else {
							pos = n
							clientCost = client.LastPosition - newOrder.Floor
						}
					} else {

					}
				}else if newOrder.Internal{
					pos = n
				}else if n < len(client.Orders)-1 {
					continue
				} else {
					clientCost = int(math.Abs(float64(newOrder.Floor - client.LastPosition)))
					pos = n+1
				}
				clientCost += n

				if clientCost < bestCost {
					bestCost = clientCost
					cost = Cost{client, clientCost, pos}
					println("current clientCost: " + strconv.Itoa(clientCost))
					println("bestCost: " + strconv.Itoa(bestCost))
				}
				break
			}
		} else {
			clientCost = int(math.Abs(float64(newOrder.Floor - client.LastPosition)))
			if clientCost < bestCost {
				bestCost = clientCost
				cost = Cost{client, clientCost, 0}
				println("current clientCost: " + strconv.Itoa(clientCost))
				println("bestCost: " + strconv.Itoa(bestCost))
			}
		}
	}

	newOrders := make([]*com.Order, 0, maxOrderSize)
	println("Plassering i ordrekø: " + strconv.Itoa(cost.OrderPos))
	if cost.OrderPos > 0 {
		newOrders = append(newOrders, cost.Client.Orders[0:cost.OrderPos]...)
		newOrders = append(newOrders, newOrder)
		newOrders = append(newOrders, cost.Client.Orders[cost.OrderPos:]...)
	} else {
		newOrders = append(newOrders, newOrder)
		newOrders = append(newOrders, cost.Client.Orders...)
	}
	bestClient := Client{cost.Client.ID, cost.Client.Active, cost.Client.IsMaster, cost.Client.LastPosition, cost.Client.Direction, nil}
	bestClient.Orders = newOrders
	println("Størrelse på ordrekø: " + strconv.Itoa(len(bestClient.Orders)))
	return bestClient
}

//func calc(newOrder *com.Order) Client {
//	//return Client{clients[0].ID,clients[0].Active,clients[0].IsMaster,clients[0].LastPosition,clients[0].Direction,clients[0].Orders}
//	const stopCost int = 1
//	var cost Cost
//	var bestCost int
//	var clientCost int
//	bestCost = 1000

//	for _, client := range clients {
//		if (client.LastPosition < 0) && (client.Direction < 0) || !client.Active {
//			continue
//		} else if len(client.Orders) > 0 {

//			println("er i kostfunksjon")
//			for n, order := range client.Orders {
//				if newOrder.Direction == order.Direction {
//					if newOrder.Floor > order.Floor && newOrder.Floor > client.LastPosition && client.LastPosition > order.Floor { // case 12,15
//						if order.Direction == 0 {
//							clientCost = client.LastPosition - 2*order.Floor + newOrder.Floor + stopCost
//						} else {
//							clientCost = client.LastPosition + newOrder.Floor + stopCost
//						}
//					} else if newOrder.Floor <= order.Floor && newOrder.Floor < client.LastPosition && client.LastPosition < order.Floor { // case 9,14
//						if order.Direction == 0 {
//							clientCost = -client.LastPosition + 2*nFloors - newOrder.Floor + 2*stopCost
//						} else {
//							clientCost = 2*order.Floor - client.LastPosition - newOrder.Floor + stopCost
//						}
//					} else if newOrder.Floor <= order.Floor && newOrder.Floor > client.LastPosition { // case 1,7
//						if order.Direction == 0 {
//							clientCost = newOrder.Floor - client.LastPosition + stopCost
//						} else {
//							clientCost = 2*order.Floor - client.LastPosition - newOrder.Floor
//						}
//					} else if newOrder.Floor <= order.Floor && newOrder.Floor < client.LastPosition { // case 3,5
//						if order.Direction == 0 {
//							clientCost = client.LastPosition - newOrder.Floor
//						} else {
//							clientCost = client.LastPosition - newOrder.Floor + stopCost
//						}
//					} else if newOrder.Floor > client.LastPosition { // case 2,8
//						if order.Direction == 0 {
//							clientCost = newOrder.Floor - client.LastPosition + stopCost
//						} else {
//							clientCost = newOrder.Floor - client.LastPosition
//						}
//					} else if newOrder.Floor < client.LastPosition { // case 4,6
//						if order.Direction == 0 {
//							clientCost = client.LastPosition - 2*order.Floor + newOrder.Floor + stopCost
//						} else {
//							clientCost = client.LastPosition - newOrder.Floor
//						}
//					} else {

//					}
//				} else if n < len(client.Orders)-1 {
//					continue
//				} else {
//					clientCost = int(math.Abs(float64(newOrder.Floor - client.LastPosition)))
//				}
//				clientCost += n

//				if clientCost < bestCost {
//					bestCost = clientCost
//					cost = Cost{client, clientCost, n}
//					println("current clientCost: " + strconv.Itoa(clientCost))
//					println("bestCost: " + strconv.Itoa(bestCost))
//				}
//			}
//		} else {
//			clientCost = int(math.Abs(float64(newOrder.Floor - client.LastPosition)))
//			if clientCost < bestCost {
//				bestCost = clientCost
//				cost = Cost{client, clientCost, 0}
//				println("current clientCost: " + strconv.Itoa(clientCost))
//				println("bestCost: " + strconv.Itoa(bestCost))
//			}
//		}
//	}

//	newOrders := make([]*com.Order, 0, maxOrderSize)
//	println("Plassering i ordrekø: " + strconv.Itoa(cost.OrderPos))
//	println("Størrlse på ordrekø: " + strconv.Itoa(len(cost.Client.Orders)))
//	if cost.OrderPos > 0 {
//		newOrders = append(newOrders, cost.Client.Orders[0:cost.OrderPos]...)
//		newOrders = append(newOrders, newOrder)
//		newOrders = append(newOrders, cost.Client.Orders[cost.OrderPos:]...)
//	} else {
//		newOrders = append(newOrders, newOrder)
//		newOrders = append(newOrders, cost.Client.Orders...)
//	}
//	bestClient := Client{cost.Client.ID, cost.Client.Active, cost.Client.IsMaster, cost.Client.LastPosition, cost.Client.Direction, nil}
//	bestClient.Orders = newOrders
//	println("Størrlse på ordrekø: " + strconv.Itoa(len(bestClient.Orders)))
//	return bestClient
//}