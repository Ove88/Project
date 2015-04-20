package cntr

var (
	clients []*Client
)

const nFloors int = 3

type Client struct {
	ID           int
	Active       bool
	IsMaster     bool
	LastPosition int
	Direction    int
	activeOrder  Order
}

type Order struct {
	MessageID int
	SendID    int
	RecvID    int
	Internal  bool
	Floor     int
	Direction int
}

func calculateCost(order Order, client Client) Client {
	var bestClient Client
	var bestCost int
	bestCost = 1000
	var clientCost int

	for _, client := range clients {
		if (client.LastPosition < 0) && (client.Direction < 0) {
			continue
		} else {
			if order.Direction == client.Direction {
				if order.Floor <= client.activeOrder.Floor {
					if order.Floor > client.LastPosition {
						clientCost = clientCost + 0
					} else if order.Floor < client.LastPosition {
						if client.activeOrder.Direction == 0 {
							clientCost = clientCost + (client.activeOrder.Floor - order.Floor) + (nFloors - client.activeOrder.Floor) + (nFloors - order.Floor)
						} else if client.activeOrder.Direction == 1 {
							clientCost = clientCost + (order.activeOrder.Floor - order.Floor) + order.activeOrder.Floor + order.Floor
						}
					}
				} else if order.Floor > client.activeOrder.Floor {
					clientCost = clientcost + (client.activeOrder.Floor - client.LastPosition) + (order.Floor - client.activeOrder.Floor)
				}
			} else { // order.Direction != client.Direction
				if client.activeOrder.Direction == 0 {
					clientCost = clientCost + (client.activeOrder.Floor - client.LastPosition) + client.activeOrder.Floor + order.Floor
				} else {
					clientCost = clientCost + (client.LastPosition) + order.Floor
				}
			}
		}

		if clientCost < bestCost {
			bestClient = client
		}
	}
	return bestClient
}
