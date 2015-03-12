package tcp

import (
	"errors"
	"net"
	"strconv"
	"strings"
	//"time"
)

type Protocol interface {
	Decode(buffer []byte) (message IDable, received bool)
	Encode(message IDable) (buffer []byte)
}

type NewProtocol interface {
	NewProtocol() (pr Protocol)
	GetBufferSize() int
}

type IDable interface {
	RemoteID() int
}

type client struct {
	active bool
	id     int
	conn   *net.TCPConn
}

type ClientStatus struct {
	ID     int
	Active bool
}

var (
	laddr      *net.TCPAddr
	raddr      *net.TCPAddr
	clients    []*client
	cStatus_ch chan ClientStatus
)

func StartServer(localIPAddr string, send_ch <-chan IDable, receive_ch chan<- interface{},
	status_ch chan ClientStatus, newpr NewProtocol, maxNumberOfClients int) (err error) {

	if newpr == nil {
		return errors.New("Protocol")
	}
	cStatus_ch = status_ch
	clients = make([]*client, 0, maxNumberOfClients)
	laddr, err = net.ResolveTCPAddr("tcp4", localIPAddr)
	if err != nil {
		return
	}

	listenConn, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		listenConn.Close()
		return
	}
	go listenForClients(listenConn, receive_ch, newpr)
	go sendPackets(send_ch, newpr.NewProtocol())
	return
}

func StartClient(localIPAddr, remoteAddr string, send_ch <-chan IDable,
	receive_ch chan<- interface{}, status_ch chan ClientStatus, newpr NewProtocol) (err error) {

	if newpr == nil {
		return errors.New("Protocol")
	}
	cStatus_ch = status_ch
	clients = make([]*client, 0, 1)
	laddr, err = net.ResolveTCPAddr("tcp4", localIPAddr)
	raddr, err = net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return
	}
	//go func() {
	conn, err := net.DialTCP("tcp4", laddr, raddr)
	if err != nil {
		println(err.Error())
		conn.Close()
		return
	}

	client_ := client{true, getClientId(conn), conn}
	clients = append(clients, &client_)

	go sendPackets(send_ch, newpr.NewProtocol())
	go receivePackets(&client_, receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
	//}()
	return
}

func listenForClients(listenConn *net.TCPListener, receive_ch chan<- interface{}, newpr NewProtocol) {

	var clientExists bool
	for {
		clientExists = false
		conn, err := listenConn.AcceptTCP()
		if err != nil {
			println(err.Error())
		}
		id := getClientId(conn)
		for i, client := range clients {
			if id == client.id {
				clients[i].conn = conn
				clients[i].active = true
				clientExists = true
				go receivePackets(clients[i], receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
				cStatus_ch <- ClientStatus{clients[i].id, clients[i].active}
				break
			}
		}
		if !clientExists {
			client_ := client{true, getClientId(conn), conn}
			clients = append(clients, &client_)
			go receivePackets(&client_, receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
			cStatus_ch <- ClientStatus{client_.id, client_.active}

		}
	}
}

func sendPackets(send_ch <-chan IDable, pr Protocol) {
	for {
		message, ok := <-send_ch
		if !ok {
			break
		}
		for i, client := range clients {
			if client.id == message.RemoteID() && clients[i].active {
				client.conn.Write(pr.Encode(message))
				break
			}
		}
	}
}

func receivePackets(client_ *client, receive_ch chan<- interface{}, pr Protocol, buffersize int) {
	nTries := 0
	buffer := make([]byte, buffersize)
	for {
		n, err := client_.conn.Read(buffer)
		if err != nil {

			if nTries > 9 {
				client_.active = false
				client_.conn.Close()
				cStatus_ch <- ClientStatus{client_.id, client_.active}
				break
			}
			nTries++
			continue
		}
		nTries = 0

		message, recv := pr.Decode(buffer[0:n])
		if recv {
			receive_ch <- message
		} else {
			continue
		}
	}
}

func getClientId(conn *net.TCPConn) (id int) {

	raddr := conn.RemoteAddr().String()
	splitString := strings.Split(raddr, ".")
	splitString = strings.Split(splitString[len(splitString)-1], ":")
	id, _ = strconv.Atoi(splitString[0])
	return
}
