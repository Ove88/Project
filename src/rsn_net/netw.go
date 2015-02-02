package rsn_net

import (
	"net"
	"tcp"
	"udp"
)

type Netw_clientData struct {
	Client_ID         int
	Elev_state        int
	Elev_position     int
	Elev_direction    int
	Elev_destinations []int
}

type Netw_initParameters struct {
	LocalID             int
	LocalIPaddress      string
	RemoteIPaddress     string
	LocalListeningport  int
	RemoteListeningport int
}

type Netw_elevOrder struct {
	Client_ID        int
	Elev_destination int
}

func Netw_masterInit(parameters Netw_initParameters, newClient_ch chan int, receiveData_ch chan Netw_clientData, receiveOrder_ch chan Netw_elevOrder) {

}

func Netw_slaveInit(parameters Netw_initParameters, sendOrder_ch chan Netw_elevOrder) {

}
