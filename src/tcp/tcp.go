package netw

import (
	//"fmt"
	"net"
	"strconv"
)

var ( 
	localaddr *net.TCPAddr
 	remoteaddr *net.TCPAddr
 	clients []client
)

type client struct {
	id  int
	conn *net.TCPConn	
}

type TcpPacket struct {
	Client_id int 
	Buffer 	  []byte
	Nbytes	  int
}

func StartListen(localIPAddr string, localPort int, packet_ch chan TcpPacket) {
	
	clients = make([]client, 5)
	localaddr = net.ResolveTCPAddr("tcp4", localIPAddr)
	localaddr.Port = localPort
	tcpListenConn := net.ListenTCP("tcp4", localaddr)

	go listen(tcpListenConn,packet_ch)
	//go receive()
	go send(packet_ch)
}

func Connect(localIPAddr, remoteIPAddr string, localPort, remotePort int, packet_ch chan TcpPacket) {

	localaddr = net.ResolveTCPAddr("tcp4", localIPAddr + ":" + strconv.Itoa(localPort))
	remoteaddr = net.ResolveTCPAddr("tcp4", remoteIPAddr + ":" + strconv.Itoa(remotePort))
	
	tcpConn := net.DialTCP("tcp4", localaddr, remoteaddr)
	client := client{0, tcpConn}
	clients = append(clients, client)
	
	//go receive()
	go send(packet_ch)
}

func send(packet_ch chan TcpPacket) {
	for {
		packet := <- packet_ch
		for _, client := range clients {
			if client.id == packet.Client_id {
				n := client.conn.Write(packet.Buffer)

				break
			}
		}
	}
}

func listen(listenConn *net.TCPListener,packet_ch chan TcpPacket) {
	
	for {
		conn := listenConn.AcceptTCP()
		client := client{0, conn}
		//go receive(client)
		//
		clients = append(clients, client)
		go receive(client,packet_ch)
	}
}

func receive(client client, packet_ch chan TcpPacket) {
	buffer := make([]byte,4096)
	for {
		n := client.conn.Read(buffer)
		packet := TcpPacket{0,buffer[0:n],n}
		packet_ch <- packet
    	}
}
