package netw

//Protocol
import (
	"encoding/json"
	"net"
	"tcp"
)

type InitParameters struct {
	LocalID             int
	LocalIPaddress      string
	RemoteIPaddress     string
	LocalListeningport  int
	RemoteListeningport int
}

type ElevData struct {
	Message_nr        int
	Client_id         int
	Elev_state        int
	Elev_position     int
	Elev_direction    int
	Elev_destinations []int
}

type ElevOrder struct {
	Client_id        int
	Elev_destination int
}

func InitMaster(parameters InitParameters,
	receiveData_ch chan ElevData, order_ch chan ElevOrder) {

	packet_ch := make(chan tcp.TcpPacket, 20)
	tcp.StartListen(parameters.LocalIPaddress, parameters.LocalListeningport, packet_ch)
	go createPacket(packet_ch)
	go getMessage(packet_ch)
}

func InitSlave(parameters InitParameters,
	order_ch chan ElevOrder, sendData_ch chan ElevData) {
}

func createPacket(packet_ch chan tcp.TcpPacket) {
	for {
		encodeMessage()
		wrapMessage()
		//packet_ch <-
	}
}

func getMessage(packet_ch chan tcp.TcpPacket) {
	for {
		packet := <-packet_ch
		unwrapMessage()
		decodeMessage()
	}
}

func wrapMessage() {

}

func unwrapMessage() {

}

func encodeMessage() {
	for {
		select {
		case data := <-data_ch:
		case order := <-order_ch:
		}
	}
}

func decodeMessage() {

}
