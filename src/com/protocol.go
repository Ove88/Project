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
)

type headerProtocol struct {
	tempBuffer []byte
}

type NewHeaderProtocol struct {
	Buffersize int
}

func (pr headerProtocol) Decode(buffer []byte) (tcp.IDable, bool) {

	pr.tempBuffer = append(pr.tempBuffer, buffer...)
	data, typeOfMessage, received := pr.unwrapMessage()
	if received {
		switch typeOfMessage {
		case "ElevData":
			var message ElevData
			json.Unmarshal(data, &message)
			return message, received
		}
		//// Legg til ny case for hver nye datastruct her ////
	}
	return nil, received
}

func (pr headerProtocol) Encode(message tcp.IDable) []byte {
	data, _ := json.Marshal(message)
	typeOfMessage := strings.Split(reflect.TypeOf(message).String(), ".")[1]
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
