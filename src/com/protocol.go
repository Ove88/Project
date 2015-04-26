package com

import (
	"com/tcp"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

const (
	protocolType     string = "header"
	prefixBufferSize int    = 5
	prefixTypeSize   int    = 20
	//useMessageHeader bool   = true
)

type headerProtocol struct {
	tempBuffer []byte
}

type NewHeaderProtocol struct {
	Buffersize int
}

type Header struct {
	MessageID int
	SendID    int
	RecvID    int
	Data      interface{}
}

func (h Header) RemoteID() int {
	return h.RecvID
}
func (h Header) GetType() string {
	return strings.Split(reflect.TypeOf(h.Data).String(), ".")[1]
}

type Order struct {
	OriginID  int
	Internal  bool
	Floor     int
	Direction int
	Cost      int
}

func (o Order) String() string {
	return "Order:" + strconv.Itoa(o.Direction) + "," + strconv.Itoa(o.Floor)
}

type Orders struct {
	ClientID int
	Orders   []*Order
}

type ElevUpdate struct {
	LastPosition int
	Direction    int
}

func (e ElevUpdate) String() string {
	return "ElevUpdate:" + strconv.Itoa(e.Direction) + "," + strconv.Itoa(e.LastPosition)
}

type Ack struct {
	Flag bool
}

type ButtonLamp struct {
	Button int
	Floor  int
	State  bool
}

/////   Sett inn flere datastructer her   /////

func (pr headerProtocol) Decode(buffer []byte) (interface{}, bool) {

	pr.tempBuffer = append(pr.tempBuffer, buffer...)
	rawMessage, typeOfMessage, received := pr.unwrapMessage()

	if received {
		if typeOfMessage == "PollMessage" {
			var data tcp.PollMessage
			json.Unmarshal(rawMessage, &data)
			return data, received
		} else {			
		var message Header
		json.Unmarshal(rawMessage, &message)
		rawMessage, _ = json.Marshal(message.Data)
		switch typeOfMessage {
		case "ElevUpdate":
		case "Position":
			var data ElevUpdate
			json.Unmarshal(rawMessage, &data)
			message.Data = data
			return message, received
		case "Order":
			var data Order
			json.Unmarshal(rawMessage, &data)
			message.Data = data
			return message, received
		case "Orders":
			var data Orders
			json.Unmarshal(rawMessage, &data)
			message.Data = data
			return message, received
		case "ButtonLamp":
			var data ButtonLamp
			json.Unmarshal(rawMessage, &data)
			message.Data = data
			return message, received
		case "Ack":
			var data Ack
			json.Unmarshal(rawMessage, &data)
			message.Data = data
			return message, received
			}
		}
		//// Legg til ny case for hver nye datastruct her ////
	}
	return nil, received
}

func (pr headerProtocol) Encode(message tcp.IDable) []byte {
	data, _ := json.Marshal(message)
	typeOfMessage := message.GetType() //strings.Split(reflect.TypeOf(message).String(), ".")[1]
	return pr.wrapMessage(data, typeOfMessage)
}

// func (pr headerProtocol) SetBufferSize(size int) {
// 	capasity := cap(pr.tempBuffer)
// 	if size > capasity {
// 		newBuffer := make([]byte, len(pr.tempBuffer), size)
// 		copy(newBuffer, pr.tempBuffer)
// 		pr.tempBuffer = newBuffer
// 	} else {
// 		pr.tempBuffer = pr.tempBuffer[0:size]
// 	}
// 	return
// }

func (pr headerProtocol) wrapMessage(data []byte, typeOfMessage string) []byte {

	headerSize := len(protocolType) + prefixTypeSize + prefixBufferSize
	header := make([]byte, 0, headerSize)

	sizeBytes := make([]byte, prefixBufferSize)
	copy(sizeBytes, []byte(strconv.Itoa(len(data))+":"))

	typeBytes := make([]byte, prefixTypeSize)
	copy(typeBytes, []byte(typeOfMessage+":"))

	header = append(header, []byte(protocolType)...)
	header = append(header, sizeBytes...)
	header = append(header, typeBytes...)
	return append(header, data...)
}

func (pr headerProtocol) unwrapMessage() (data []byte, typeOfMessage string, received bool) {

	headerSize := len(protocolType) + prefixTypeSize + prefixBufferSize
	header := pr.tempBuffer[0:headerSize]
	//protocoltype := string(header[0:len(protocolType)])

	buffersize, _ := strconv.Atoi(
		strings.Split(string(header[len(protocolType):len(protocolType)+prefixBufferSize]), ":")[0])

	typeOfMessage = strings.Split(
		string(header[headerSize-prefixTypeSize:headerSize]), ":")[0]

	if buffersize > len(pr.tempBuffer)-headerSize {
		received = false
		return
	} else {
		received = true
		data = make([]byte, buffersize)

		copy(data, pr.tempBuffer[headerSize:headerSize+buffersize])
		pr.tempBuffer = pr.tempBuffer[headerSize+buffersize:]
		return
	}
}

func (npr NewHeaderProtocol) NewProtocol() (protocol tcp.Protocol) {
	return headerProtocol{make([]byte, 0, npr.Buffersize)}
}

func (npr NewHeaderProtocol) GetBufferSize() int {
	return npr.Buffersize
}
