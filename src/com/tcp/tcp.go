package tcp

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"reflect"
	"time"
)

type Protocol interface {
	Decode(buffer []byte) (message interface{}, received bool)
	Encode(message IDable) (buffer []byte)
}

type NewProtocol interface {
	NewProtocol() (pr Protocol)
	GetBufferSize() int
}

type IDable interface {
	RemoteID() int
	GetType() string
}

type client struct {
	active bool
	id     int
	conn   *net.TCPConn
	netTimer *time.Timer
}

type ClientStatus struct {
	ID       int
	Active   bool
	IsMaster bool
}

func (c ClientStatus) String() string {
	return strconv.Itoa(c.ID) + ":" + strconv.FormatBool(c.Active)
}

type PollMessage struct{
	recvID 	int
	message string
}

func (p PollMessage) GetType() string{
	return strings.Split(reflect.TypeOf(p).String(), ".")[1]
}

func (p PollMessage) RemoteID() int{
	return p.recvID
}

var (
	active  bool
	laddr      *net.TCPAddr
	raddr      *net.TCPAddr
	listenConn *net.TCPListener
	clients    []*client
	cStatus_ch chan ClientStatus
)

func StartServer(localIPAddr string, send_ch <-chan IDable, receive_ch chan<- interface{},
	status_ch chan ClientStatus, newpr NewProtocol, maxNumberOfClients int) (masterPort int, err error) {

	if newpr == nil {
		return 0, errors.New("Protocol")
	}
	cStatus_ch = status_ch
	clients = make([]*client, 0, maxNumberOfClients)
	laddr, err = net.ResolveTCPAddr("tcp4", localIPAddr+":0")
	if err != nil {
		return
	}

	listenConn, err = net.ListenTCP("tcp4", laddr)
	if err != nil {
		listenConn.Close()
		return
	}
	masterPort, _ = strconv.Atoi(strings.SplitAfterN(listenConn.Addr().String(), ":", 2)[1])

	go listenForClients(listenConn, receive_ch, newpr)
	go sendPackets(send_ch, newpr.NewProtocol())
	go pollClients(newpr.NewProtocol())
	return
}

func StartClient(localIPAddr, remoteAddr string, send_ch <-chan IDable,
	receive_ch chan<- interface{}, status_ch chan ClientStatus, newpr NewProtocol) (err error) {

	cStatus_ch = status_ch
	clients = make([]*client, 0, 1)
	laddr, err = net.ResolveTCPAddr("tcp4", localIPAddr+":0")
	raddr, err = net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return
	}

	conn, err := net.DialTCP("tcp4", laddr, raddr)
	if err != nil {
		println(err.Error())
		conn.Close()
		return
	}
	println("starter client")
	client_ := client{true, getClientId(conn), conn,time.NewTimer(2*time.Second)}
	clients = append(clients, &client_)

	go sendPackets(send_ch, newpr.NewProtocol())
	go receivePackets(&client_, receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
	go masterConnTimeoutListener()
	
	cStatus_ch <- ClientStatus{client_.id, client_.active, true}
	return
}

func listenForClients(listenConn *net.TCPListener, receive_ch chan<- interface{}, newpr NewProtocol) {

	var clientExists bool
	for {
		clientExists = false
		conn, err := listenConn.AcceptTCP()
		if err != nil {
			println("Closing listenconn")
			active = false
			listenConn.Close()
			cStatus_ch <- ClientStatus{-1, false, false}
			break
		}
		id := getClientId(conn)
		for i, client := range clients {
			if id == client.id {
				clients[i].conn = conn
				clients[i].active = true
				clientExists = true
				go receivePackets(clients[i], receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
				cStatus_ch <- ClientStatus{clients[i].id, clients[i].active, false}
				break
			}
		}
		if !clientExists {
			client_ := client{true, getClientId(conn), conn,time.NewTimer(2*time.Second)}
			clients = append(clients, &client_)
			go receivePackets(&client_, receive_ch, newpr.NewProtocol(), newpr.GetBufferSize())
			cStatus_ch <- ClientStatus{client_.id, client_.active, false}
		}
	}
}

func sendPackets(send_ch <-chan IDable, pr Protocol) {
	active = true
	for active{
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
	
	buffer := make([]byte, buffersize)
	for {
		n, err := client_.conn.Read(buffer)
		if err != nil {
			println("closing clientconn")
			client_.active = false
			client_.conn.Close()
			cStatus_ch <- ClientStatus{client_.id, client_.active, false}
			break
		}
		message, recv := pr.Decode(buffer[0:n])
		if recv {
			switch message.(type){
				case PollMessage:
					println("receiving Pollmessage")
					if !client_.netTimer.Reset(time.Second){
						client_.netTimer = time.NewTimer(time.Second)
					}
					client_.conn.Write(pr.Encode(PollMessage{0,"keepAlive"}))
				default:
					receive_ch <- message
			}		
		} else {
			continue
		}
	}
}

func masterConnTimeoutListener(){
	active = true
	for active{
		select{
			case <-time.After(100*time.Millisecond):
				continue
			case <- clients[0].netTimer.C:
				clients[0].conn.Close()
				clients[0].active = false
				active = false	
		}
	}
}

func pollClients(pr Protocol) {
	var noClients bool
	active = true
	for active {
		noClients = true
		for n,client := range clients {	
			select{
				case <-time.After(10*time.Millisecond):
					if client.active {
						noClients = false
						println("sending Pollmessage")
						client.conn.Write(pr.Encode(PollMessage{0,"keepAlive"}))
					}
				case <-client.netTimer.C:
					clients[n].conn.Close()
					clients[n].active = false
			}
		}
		if noClients{
			active = false
			listenConn.Close()
		}
		time.Sleep(200*time.Millisecond)
	}
}

func getClientId(conn *net.TCPConn) (id int) {

	raddr := conn.RemoteAddr().String()
	splitString := strings.Split(raddr, ".")
	splitString = strings.Split(splitString[len(splitString)-1], ":")
	id, _ = strconv.Atoi(splitString[0])
	return
}
