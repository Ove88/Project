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
		if (client.LastPosition < 0) && (client.Direction < 0) || !client.Active {
			continue
		} else if len(client.
			for n, order := range client.Orders {
			if order.Direction == client.Direction {
				if order.Floor <= client.activeOrder.Floor {
					if order.Floor > client.LastPosition {
						clientCost = 0
					} else if order.Floor < client.LastPosition {
						if client.activeOrder.Direction == 0 {
							clientCost =   2*nFloors - 2*order.Floor
						} else if client.activeOrder.Direction == 1 {
							clientCost = 2* order.activeOrder.Floor
						}
					}
				} else if order.Floor > client.activeOrder.Floor {
					clientCost = order.Floor - client.LastPosition
				}
			} else { // order.Direction != client.Direction
				if client.activeOrder.Direction == 0 {
					clientCost = 2*client.activeOrder.Floor + order.Floor - client.LastPosition  
				} else {
					clientCost = client.LastPosition + order.Floor
				}
			}
		}

		if clientCost < bestCost {
			bestClient = client
		}
	}
	return bestClient
}
