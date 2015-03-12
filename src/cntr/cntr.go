package cntr

type Client struct {
	ID        int
	Position  int
	Direction int
	Orders    []Order
}

type Order struct {
	ID        int
	Position  int
	Direction int
}

var clients []Client

func c(o Order) {
	for i, client := range clients {
		for j, order := range client.Orders {

		}
	}

}
