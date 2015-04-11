package com

import (
	"strconv"
)

type ElevUpdate struct {
	MessageID 	int
	SendID   	int
	RecvID    	int
	LastPosition int
	Direction 	int
}

func (e ElevUpdate) RemoteID() int {
	return e.RecvID
}
func (e ElevUpdate) String() string {
	return "TID:" + strconv.Itoa(e.SendID) + ", RecvID:" +
		strconv.Itoa(e.RecvID) + ", State:" + e.Direction
}

type Order struct {
	MessageID 	int
	SendID   	int
	RecvID    	int
	Internal 	bool
	Floor     	int
	Direction 	int
}

type Ack struct {
	MessageID 	int
	SendID 		int
	RecvID  		int
}

type Orders struct {
	MessageID 	int
	SendID 		int
	RecvID  		int
	Orders  		*[]Order
}

/////   Sett inn flere datastructer her   /////
