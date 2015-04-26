package main

import (
	"com"
	"com/tcp"
	"elevator"
	"math"
	"strconv"
	"time"
	//"reflect"
)

//TODO Flere ordre på samme knapp

const (
	maxOrderSize       int = 50
	maxNumberOfClients int = 10
	nFloors            int = 4
	activityTimeout    int = 4
)

var (
	send_ch             chan tcp.IDable
	receive_ch          chan interface{}
	clientStatus_ch     chan tcp.ClientStatus
	lOrderReceive_ch    chan elevator.Order
	lOrderSend_ch       chan elevator.Order
	elevPos_ch          chan elevator.Position
	message_ch          chan com.Header
	ack_ch              chan com.Header
	elevUpdate_ch       chan com.Header
	reCalc_ch           chan *com.Order
	stopBtn_ch          chan bool
	clients             []*Client
	masterID            int
	clientNumber 		int
	isAlone             bool
	elevPositionChanged bool
)

type Client struct {
	ID            int
	Active        bool
	IsMaster      bool
	LastPosition  int
	Direction     int
	Orders        []*com.Order
	ActivityTimer *time.Timer
}

type Cost struct {
	Client   *Client
	Cost     int
	OrderPos int
}

func main() {
	masterID = -1
	clientNumber = 0
	isAlone = true
	clients = make([]*Client, 0, maxNumberOfClients)
	send_ch = make(chan tcp.IDable, 1)
	receive_ch = make(chan interface{}, 10)
	clientStatus_ch = make(chan tcp.ClientStatus, 5)
	lOrderSend_ch = make(chan elevator.Order, 5)
	lOrderReceive_ch = make(chan elevator.Order, 5)
	elevPos_ch = make(chan elevator.Position, 5)
	ack_ch = make(chan com.Header, 50)
	message_ch = make(chan com.Header, 50)
	elevUpdate_ch = make(chan com.Header, 100)
	reCalc_ch = make(chan *com.Order, 100)
	stopBtn_ch = make(chan bool, 1)

	if !elevator.Init(lOrderSend_ch, lOrderReceive_ch, elevPos_ch) {
		println("Failed to initialize elevator hardware")
	}
	localID, ok := com.Init(send_ch, receive_ch, clientStatus_ch, maxNumberOfClients)
	if !ok {
		println("Failed to initialize network")
	}
	clients = append(
		clients, &Client{localID, true, true, 0, 0, make([]*com.Order, 0, maxOrderSize), time.NewTimer(10 * time.Second)})
	println(localID)

	go netwMessageHandler()
	go channelSelector()
	go activityTimersHandler()
	clientStatusManager()
}

func activityTimersHandler() {
	for {
		for n, client := range clients {
			select {
			case <-client.ActivityTimer.C: // No activity registered on client after given time
				if clients[0].IsMaster {
					if client.ID == clients[0].ID && !elevPositionChanged || client.ID != clients[0].ID{
						if client.Orders != nil && len(client.Orders) > 0{
							clients[n].Active = false
							println("client " + strconv.Itoa(client.ID) + " inaktiv")
							for i, order := range clients[n].Orders {
								if !order.Internal {
									clients[n].Orders[i].OriginID = clients[n].ID
									reCalc_ch <- client.Orders[i]
								}
							}
						}
					}
				} 
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}
}

func netwMessageHandler() {
	for {
		message_ := <-receive_ch

		switch message := message_.(type) {
		
		case com.Header:
			switch data := message.Data.(type) {
			
			case com.ButtonLamp: // Button light update from master
				println("buttonLamp")
				elevator.SetButtonLamp(
					data.Button, data.Floor, data.State)
			
			case com.Ack: // Ack message from client
				ack_ch <- message
			
			default: // All other messages
				message_ch <- message
			}
		default:
		println("ikke header")
		}
	}
}
func channelSelector() {
	for {
		select {
		case message := <-message_ch: // Messages from the network
			transactionManager(&message,false)

		case lorder := <-lOrderReceive_ch: // Local orders from elevator
			//orderExists := false
			if clients[0].Active { // Accept order only if active
				
				//for _,client:= range clients{ // Discard order if it already exists
				//	for _,order := range client.Orders{
				//		if lorder.Floor == order.Floor && 
				//			lorder.Direction == order.Direction && 
				//			lorder.Internal == order.Internal{
				//			orderExists = true
				//		}
				//	}
				//}
				//if !orderExists{
					lorder.OriginID = clients[0].ID
					transactionManager(&com.Header{
						newMessageID(), clients[0].ID, 0, lorder},false)
				//}
			}
		
		case order := <-reCalc_ch: // Recalculate orders from inactive client
			println("recalc")
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, com.Order{
				order.OriginID,order.Internal,order.Floor,order.Direction,order.Cost}},true)
		
		case pos := <-elevPos_ch: // Elevator position has changed
			transactionManager(&com.Header{
				newMessageID(), clients[0].ID, 0, pos},false)
		}
	}
}

func transactionManager(message *com.Header, recalc bool) bool {

	switch data := message.Data.(type) {

	case elevator.Order: // Local order
		orderOK := false
		for !orderOK {
			if clients[0].Active {
				if clients[0].IsMaster {
					chosenClient := calculateClient(elevToCom(&data))
					if chosenClient.ID == 0{
						orderOK = true
					}
					if orderUpdater(elevToCom(&data), &chosenClient, true) {
						for n, client_ := range clients {
							if client_.ID == chosenClient.ID {
								clients[n].Orders = chosenClient.Orders
								orderOK = true
								break
							}
						}
					} else { // No ack from client
						for n, client_ := range clients {
							if client_.ID == chosenClient.ID {
								clients[n].Active = false
								continue
							}
						}
					}
				} else {
					message.RecvID = masterID
					send_ch <- message
					orderOK = true
				}
			}
		}
	case elevator.Position: // Elevator has changed position
		clients[0].LastPosition = data.LastPosition
		clients[0].Direction = data.Direction

		if clients[0].IsMaster {
			clientNumber = 0 //fix
			if !clients[0].Active {
				clients[0].Active = true
			}
			if len(clients[0].Orders) > 0 { // Length of order queue larger than zero

				if data.LastPosition == clients[0].Orders[0].Floor &&
					data.Direction == -1 { // Elevator has reached its destination

					elevPositionChanged = true

					println("Klient " + strconv.Itoa(clients[0].ID) + " har ankommet etasje " + strconv.Itoa(clients[0].LastPosition))
					lastOrder := clients[0].Orders[0]
					clients[0].Orders = clients[0].Orders[1:]
					orderUpdater(lastOrder, clients[0], false)																					//drtxdgdrtddfg
					
					clients[0].ActivityTimer = time.NewTimer(15 * time.Second)
					
					if len(clients[0].Orders) > 0 {
						lOrderSend_ch <- comToElev(clients[0].Orders[0]) // Update elevator with next order
					} else {
						clients[0].ActivityTimer.Stop()
					}
				} else if data.LastPosition == -1 && data.Direction == -1 { // Stop button pressed(master))
					if elevPositionChanged {
						elevPositionChanged = false
						stopBtn_ch <- true
					}
				} else {
					elevPositionChanged = true
					clients[0].ActivityTimer = time.NewTimer(10 * time.Second)
				}
			} else {
				elevPositionChanged = true
				clients[0].ActivityTimer = time.NewTimer(10 * time.Second)
			}

		} else if data.LastPosition == -1 && data.Direction == -1 { // Stop button pressed(slave)
			if elevPositionChanged {
				elevPositionChanged = false
				stopBtn_ch <- true
			}
		} else if clients[0].Active { // Update to master
			clientNumber = 0 //fix
			elevPositionChanged = true
			message.RecvID = masterID
			send_ch <- message
			
		} else if len(clients[0].Orders) > 0 { // Client not active
			println("clientNumber:"+strconv.Itoa(clientNumber))
			if data.LastPosition == clients[clientNumber].Orders[0].Floor &&
				data.Direction == -1 { 								// Elevator has reached its destination and not active
				println("Klient " + strconv.Itoa(clients[0].ID) + " har ankommet etasje " + strconv.Itoa(clients[0].LastPosition))
				clients[clientNumber].Orders = clients[clientNumber].Orders[1:]
				//if len(clients[0].Orders) > 0 {
				//	lOrderSend_ch <- comToElev(clients[0].Orders[0]
				//}
				println("er her 1")
				for{
					if len(clients[clientNumber].Orders) > 0 {	
						println("er her inneee")
						lOrderSend_ch <- comToElev(clients[clientNumber].Orders[0]) // Update elevator with next order	
						break			
					}else if len(clients) > clientNumber+1 {
						println("er her 2")		
						clientNumber+= 1
						if len(clients[clientNumber].Orders)>0{
							lOrderSend_ch <- comToElev(clients[clientNumber].Orders[0])
							println("er her 3")	
							break
						}
					}else{
						println("er her 4hj")	
						break
					}	
				}
			
			}
		}

	case com.Order: // Remote order from client, or recalc
	
		orderOK := false
		for !orderOK {
			chosenClient := calculateClient(&data)
			if chosenClient.ID == 0{
				orderOK = true
			}
			if orderUpdater(&data, &chosenClient, true) {
				for n, client_ := range clients {
					if recalc && clients[n].ID == data.OriginID {
						for i,order := range clients[n].Orders{
							if order.Floor == data.Floor && 
								order.Direction == data.Direction { // Removes recalculated order from last queue
								if len(clients[n].Orders) < i{
									clients[n].Orders = append(
									clients[n].Orders[0:i],clients[n].Orders[i+1:]...)
								}else{
									clients[n].Orders = nil
								}
								orderOK = true	
							}								
						}					
					}
					if client_.ID == chosenClient.ID {
						clients[n].Orders = chosenClient.Orders	
						orderOK = true					
					}
				}
			} else { // No ack from client
				for n, client_ := range clients {
					if client_.ID == chosenClient.ID {
						clients[n].Active = false
						continue
					}
				}
			}
		}
	case com.ElevUpdate: // Status update from client
		for i, client := range clients {

			if message.SendID == client.ID { // Updates elevator position
				clients[i].LastPosition = data.LastPosition
				clients[i].Direction = data.Direction

				if !client.Active {
					clients[i].Active = true
				}
				if len(client.Orders) > 0 { // Length of order queue larger than zero
					
					if data.LastPosition == clients[i].Orders[0].Floor &&
						data.Direction == -1 { // Elevator has reached its destination

						println("---Klient " + strconv.Itoa(client.ID) + " har ankommet etasje " + strconv.Itoa(client.LastPosition)+"---")
						lastOrder := clients[i].Orders[0]
						clients[i].Orders = clients[i].Orders[1:]
						orderUpdater(lastOrder, clients[i], false)
						clients[i].ActivityTimer = time.NewTimer(8 * time.Second)

						if len(clients[i].Orders) < 1 {
							clients[i].ActivityTimer.Stop()
						}
					} else {
						clients[i].ActivityTimer = time.NewTimer(4 * time.Second)
					}
				}
				break
			}
		}
	case com.Orders: // Order updates
		clientExists := false
		if !clients[0].IsMaster {
			for i, client := range clients {
				if client.ID == data.ClientID {
					//if len(data.Orders) == 0 && len(client.Orders)>0{
					//	message.RecvID = message.SendID
					//	message.SendID = clients[0].ID
					//	message.Data = client.Orders
					//	send_ch<-message
					//	buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}
					//	for _, order := range client.Orders { // Set correct button lights
					//		if !order.Internal {
					//			buttonLightUpdate.Data = com.ButtonLamp{order.Direction, order.Floor, true}
					//			send_ch <- buttonLightUpdate
					//		}
					//	}
					//}else{
						clients[i].Orders = data.Orders
						clientExists = true
					}
					if i == 0 && len(clients[0].Orders) > 0 { // Order for this elevator
						lOrderSend_ch <- comToElev(clients[0].Orders[0]) // Update elevator with order, even if no change

						send_ch <- com.Header{
							message.MessageID, clients[0].ID, message.SendID, com.Ack{true}} // Ack to master
					}
					return true
				}
			
			if !clientExists { //Create new client
				clients = append(clients, &Client{
					data.ClientID, false, false, 0, 0, data.Orders, time.NewTimer(10 * time.Second)})
			}
		}/*else { // If master
			for i, client := range clients {
				if client.ID == data.ClientID {
					clients[i].Orders = data.Orders
				}
			}
		}*/
	}
	return true
}
					//message := com.Header{newMessageID(), clients[0].ID, status.ID, nil}
					//buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}
//for _, client := range clients { // Update new client with order lists
					//if len(client.Orders) > 0 {
					//	message.Data = com.Orders{client.ID, client.Orders}
					//	send_ch <- message
					//}
					//for _, order := range client.Orders { // Set correct button lights
					//	if !order.Internal {
					//		buttonLightUpdate.Data = com.ButtonLamp{order.Direction, order.Floor, true}
					//		send_ch <- buttonLightUpdate
					//	}
					//}
				//}

func orderUpdater(order *com.Order, client *Client, isNewOrder bool) bool {
	
	orderUpdate := com.Header{
		newMessageID(), clients[0].ID, 0, com.Orders{client.ID, client.Orders}}
	
	for n, client_ := range clients { // Update all clients with new order list
		if n == 0 {
			continue
		}
		orderUpdate.RecvID = client_.ID
		send_ch <- orderUpdate
	}
	if isNewOrder && !isAlone && client.ID != clients[0].ID { // Wait for Ack if order is sent to another client
		select {
		case ack := <-ack_ch:
			if ack.MessageID != orderUpdate.MessageID || ack.SendID != client.ID {
				return false
			}
		case <-time.After(50 * time.Millisecond):
			return false
		}
	}
	if client.ID == clients[0].ID && len(client.Orders) >0 { //  Update order on this(master) elevator                          //fix
		lOrderSend_ch <- comToElev(client.Orders[0])
	}
	// Button light updates
	if !order.Internal {
		
		buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, com.ButtonLamp{
			order.Direction, order.Floor, isNewOrder}}

		for n, client_ := range clients { // Set button light for clients
			if n == 0 {
				println("sender buttonlight1:")
				elevator.SetButtonLamp(order.Direction, order.Floor, isNewOrder)
				continue
			}
			println("sender buttonlight2")
			buttonLightUpdate.RecvID = client_.ID
			send_ch <- buttonLightUpdate
		}
	} else if client.ID == clients[0].ID {
		println("sender buttonlight3")
		elevator.SetButtonLamp(2, order.Floor, isNewOrder)
	} else { 
		buttonLightUpdate := com.Header{
			newMessageID(), clients[0].ID, client.ID, com.ButtonLamp{2, order.Floor, isNewOrder}}
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
				
				if status.ID == clients[0].ID && status.Active{
					isAlone = true
					println("ØL")
					
				}else if status.ID == clients[0].ID && !client.IsMaster{
					if !status.Active {
						isAlone = true
						println("BRENNEVIN")
					}
				//}else{
				//	isAlone = false
				}
								
				if status.IsMaster {
					masterID = status.ID // Sets master ID
					println("masterID:" + strconv.Itoa(masterID))
					
					if client.ID == clients[0].ID { // Recalc orders of inactive clients after becoming master //&& !isAlone
						for g,cl := range clients {
							if !cl.Active && len(cl.Orders) > 0 {
								println("recalc inaktiv")
								for s, order := range client.Orders {
									if !order.Internal {
										clients[g].Orders[s].OriginID = clients[g].ID
										reCalc_ch <- client.Orders[s]
									}
								}
							}
						}
					}
				}
				clientExists = true
				clients[n].Active = status.Active
				clients[n].IsMaster = status.IsMaster
				
			} else if status.ID == masterID && !status.Active { // Removes master ID
				masterID = -1
			}

			if clients[0].IsMaster && clients[0].Active && n != 0 {
				isAlone = false //fix
				if status.Active { // Update existing client with order lists
					message := com.Header{newMessageID(), clients[0].ID, status.ID, nil}
					buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}

					for _, client := range clients {
						//if client.ID == status.ID { // Does not update the client with its own list //fix
						//	continue
						//}
						message.Data = com.Orders{client.ID, client.Orders}
						send_ch <- message
						for _, order := range client.Orders {
							if !order.Internal {
								buttonLightUpdate.Data = com.ButtonLamp{order.Direction, order.Floor, true}
								send_ch <- buttonLightUpdate
							}
						}
					}
				} else if clients[n].ID != clients[0].ID && !isAlone { // Recalculate orders for inactive client
					println("recalculate")
					clients[n].ActivityTimer.Stop()

					for i, order := range clients[n].Orders {
						if order.Internal{
							clients[n].Orders = append(//TODO
								clients[n].Orders[0:i],clients[n].Orders[i+1:]...) 	
						}
					}
					for i, order := range clients[n].Orders {
						if !order.Internal {
						clients[n].Orders[i].OriginID = clients[n].ID
						reCalc_ch <- client.Orders[i]
						}						
					}
				}
			}
			break
		}
		if !clientExists { // New client connected
			isAlone = false
			clients = append(clients, &Client{
				status.ID, status.Active, status.IsMaster, 0, 0, make([]*com.Order, 0, maxOrderSize), time.NewTimer(10 * time.Second)})

			if status.IsMaster {
				masterID = status.ID // Sets master ID
				println("masterID:" + strconv.Itoa(masterID))
			}
			if clients[0].IsMaster {
				message := com.Header{newMessageID(), clients[0].ID, status.ID, nil}
				buttonLightUpdate := com.Header{newMessageID(), clients[0].ID, 0, nil}

				for _, client := range clients { // Update new client with order lists
					if len(client.Orders) > 0 {
						message.Data = com.Orders{client.ID, client.Orders}
						send_ch <- message
					}
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

func elevToCom(order *elevator.Order) *com.Order { // Type converter
	return &com.Order{order.OriginID, order.Internal, order.Floor, order.Direction, 0}
}

func comToElev(order *com.Order) elevator.Order { // Type converter
	return elevator.Order{order.OriginID, order.Internal, order.Floor, order.Direction}
}

func calculateClient(newOrder *com.Order) Client {
	const stopCost int = 1
	var cost Cost
	var clientCost int
	var pos int
	var intpos int
	var tmpOrder *com.Order
	last := false
	internal := false
	first := true
	start := 0
	number := 0
	bestCost := 1000

	for _, client := range clients {
		println("*Bestilling i etasje: "+strconv.Itoa(newOrder.Floor))
		println("*****")
		println("Klient " + strconv.Itoa(client.ID)+" kalkulerer...")
		if ((client.LastPosition < 0) && (client.Direction < 0)) || !client.Active || (client.ID != newOrder.OriginID && newOrder.Internal) {
			continue
		} else if len(client.Orders) > 0 {
			for n, order := range client.Orders {
				tmpOrder = order
				if newOrder.Direction == order.Direction {
					
					if newOrder.Internal {
						if order.Internal{
							println("her 1")
							start = n+1
							continue
						}else if newOrder.Direction == 0 {
							if (newOrder.Floor > order.Floor) && (order.Floor > client.LastPosition) {
								println("her 2")
								start = n + 1
								continue
							} else if (newOrder.Floor > order.Floor) && (order.Floor == client.LastPosition) && (client.Direction == -1) {
								println("her 3")
								start = n + 1
								continue
							} else if (newOrder.Floor > order.Floor) && (newOrder.Floor > (client.LastPosition-1)){
								println("her 4")
								start = n+1
								continue
							}else{
								println("her 5")
								start = n
								break
							}
						} else {
							if (newOrder.Floor < order.Floor) && (order.Floor < (client.LastPosition)) {
								println("her 6")
								start = n + 1
								continue
							} else if (newOrder.Floor < order.Floor) && (order.Floor == client.LastPosition) && (client.Direction == -1) {
								println("her 7")
								start = n + 1
								continue
							} else if (newOrder.Floor < order.Floor) && (newOrder.Floor < (client.LastPosition+1)){
								println("her 8")
								start = n + 1
								continue
							} else{
								println("her 9")
								start = n
								break
							}

						}
					} else if order.Internal {
						if newOrder.Direction == 0 {
							if (newOrder.Floor < order.Floor) && (newOrder.Floor < (client.LastPosition + 1)) {
								println("her 10")
								start = n + 1
								continue
							} else {
								println("her 11")
								start = n
								break
							}
						} else {
							if (newOrder.Floor > order.Floor) && (newOrder.Floor > (client.LastPosition - 1)) {
								println("her 12")
								start = n + 1
								continue
							} else {
								println("her 13")
								start = n
								break
							}
						}
					}
					if !last && first {
						println("her 14")
						start = n
						number += 1
						last = true
						first = false
						continue
					} else if last && !first {
						println("her 14.5")
						number += 1
						last = true
					}
				} else if newOrder.Internal {
					if order.Internal{
						println("her 15")
						start = n+1
						continue
					}else if client.Direction == -1 {
						start = 0
						break
					} else {
						println("her 16")
						start = n + 1
						continue
					}
				} else if order.Internal {
					println("her 17")
					start = n + 1
					continue
				} else {
					if last {
						if ((newOrder.Floor <= client.LastPosition) && (newOrder.Direction == 0)) || 
							((newOrder.Floor >= client.LastPosition) && (newOrder.Direction == 1)) {
							println("her 18")
							start = n + 1
							number = 0
							last = false
							continue
						}
						break
					}
					println("her 19")
					start = n + 1
					number = 0
					continue
				}
			}
			if number != 0 { //TODO
				println("client.Orders: "+strconv.Itoa(len(client.Orders)))
				println("start: "+strconv.Itoa(start))
				println("number: "+strconv.Itoa(number))
				println("start+number: "+strconv.Itoa(start+number))
				slice:= client.Orders
				if start + number > len(client.Orders){
					slice = client.Orders[start : (len(client.Orders)-1)]
				}else{
					slice = client.Orders[start : start+number]
				}
				for place, order := range slice {
					tmpOrder = order
					if newOrder.Direction == 0 {
						if newOrder.Floor > order.Floor { //
							pos = place + 1
							continue
						} else {
							pos = place
							break
						}
					} else {
						if newOrder.Floor < order.Floor {
							pos = place + 1
							continue
						} else {
							pos = place
							break
						}
					}
				}
				intpos = pos
				pos = start + pos
				clientCost = int(math.Abs(float64(client.LastPosition-tmpOrder.Floor))) + 
					pos*stopCost + int(math.Abs(float64(newOrder.Floor-tmpOrder.Floor)))
			} else {
				pos = start
				clientCost = int(math.Abs(float64(client.LastPosition-tmpOrder.Floor))) + 
					pos*stopCost + int(math.Abs(float64(newOrder.Floor-tmpOrder.Floor)))
			}
			if clientCost < bestCost && !internal {
				bestCost = clientCost
				cost = Cost{client, bestCost, pos}
			} else if internal {
				bestCost = clientCost
				cost = Cost{client, bestCost, intpos}
			}
		} else {
			clientCost = int(math.Abs(float64(newOrder.Floor - client.LastPosition)))
			if clientCost < bestCost {
				bestCost = clientCost
				newOrder.Cost = bestCost
				cost = Cost{client, clientCost, 0}
			}
		}
		println("Klient" + strconv.Itoa(cost.Client.ID) + "'s beste kost: " + strconv.Itoa(cost.Cost))
		println("*****")
	}
	
	if bestCost == 1000 { // No active client was found 
		var empty Client
		return empty
	}
	
	newOrders := make([]*com.Order, 0, maxOrderSize)
	if cost.OrderPos > 0 {
		if internal {
			sliceOfOrders := cost.Client.Orders[start : start+number]
			sliceOfOrders = append(sliceOfOrders, cost.Client.Orders[0:start]...)
			newOrders = append(sliceOfOrders, cost.Client.Orders[start+number:]...)
		}
		newOrders = append(newOrders, cost.Client.Orders[0:cost.OrderPos]...)
		newOrders = append(newOrders, newOrder)
		newOrders = append(newOrders, cost.Client.Orders[cost.OrderPos:]...)
	} else {
		newOrders = append(newOrders, newOrder)
		if cost.Client.Orders != nil && len(cost.Client.Orders) > 0 {
			newOrders = append(newOrders, cost.Client.Orders...)
		}
	}
	bestClient := Client{cost.Client.ID, cost.Client.Active, cost.Client.IsMaster, 
		cost.Client.LastPosition, cost.Client.Direction, nil, cost.Client.ActivityTimer}
	bestClient.Orders = newOrders

	println("")
	println("---------------")
	println("Klient " + strconv.Itoa(bestClient.ID) + " tar bestillingen")
	println("Jeg har nå: " + strconv.Itoa(len(bestClient.Orders)) + " ordrer i køen")
	println("---------------")
	println("")
	return bestClient
}