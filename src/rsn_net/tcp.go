package rsn_net

import (
	"fmt"
	"net"
	"strconv"
)

var localaddr *net.TCPAddr
var clients []tcpClient

type tcpClient struct {
	ClientConn *net.TCPConn
	ID int
}

type Tcp_parameters struct {
	LocalIPaddress string
	RemoteIPaddress string 
	LocalListeningport int
	RemoteListeningport int
	NewClient_ch chan int
}
type Tcp_package struct {
	Client_ID []int
	//TODO: Implement data types
	}

func Tcp_init(parameters Tcp_parameters, myID int, ) {

	localaddr, err = net.ResolveTCPAddr("tcp4", 
	parameters.LocalIPaddress + ":" + strconv.Itoa(parameters.LocalListeningport))
	if err != nil {
		return err	
}

func Tcp_distribute(data Tcp_package, sent_ch chan bool) {

}

func Tcp_startListen(parameters Tcp_parameters) {

	tcpListenConn, err := net.ListenTCP("tcp4", localaddr)
	if err != nil {
		return err

	go tcp_listener(tcpListenConn, parameters.NewClient_ch)
}

func Tcp_startConnect(timeout int) {
	go func() {

		tcpConn, err := net.DialTCP("tcp4", laddr, raddr)

	}

}




func tcp_listener(listenConn *net.TCPConn, newClient_ch chan tcpClient) {
	
	for {
		clientConn, err := listenConn.AcceptTCP()

		client := tcpClient{clientConn}//+ ID
		clients = append(clients,client)
		
	}
}
