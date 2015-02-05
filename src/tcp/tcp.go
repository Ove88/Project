package netw

import (
	"fmt"
	"net"
	"strconv"
	"builtin"
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
	
	localaddr, err = net.ResolveTCPAddr("tcp4", localIPAddr + ":" + strconv.Itoa(localPort))
	tcpListenConn, err := net.ListenTCP("tcp4", localaddr)

	go listen(tcpListenConn)
	go receive()
	go send(packet_ch)
}

func Connect(localIPAddr, remoteIPAddr string, localPort, remotePort int, packet_ch chan TcpPacket) {

	localaddr, err = net.ResolveTCPAddr("tcp4", localIPAddr + ":" + strconv.Itoa(localPort))
	remoteaddr = net.ResolveTCPAddr("tcp4", remoteIPAddr + ":" + strconv.Itoa(remotePort))
	
	tcpConn, err := net.DialTCP("tcp4", localaddr, remoteaddr)
	client := client{0, conn}
	clients = append(clients, client)
	
	go receive()
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

func listen(listenConn *net.TCPConn) {
	
	for {
		conn, err := listenConn.AcceptTCP()
		client := client{0, conn}
		go receive(client)
		//
		clients = append(clients, client)
	}
}

func receives() {
	for _, client := range clients {
		select {
			case data := <- client.conn.Read(b)

		}
	}
}

func receive(client client, packet_ch chan TcpPacket) {
	buffer = make([]byte,4096)
	for {
		n := client.conn.Read(buffer)
		packet := TcpPacket{0,buffer[0,n],n}
		packet_ch <- packet
    }
}