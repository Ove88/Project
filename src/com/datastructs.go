package com

import (
	"strconv"
)

type ElevUpdate struct {
	TransID   int
	RecvID    int
	State     int
	Position  int
	Direction string
}

func (e ElevUpdate) RemoteID() int {
	return e.RecvID
}
func (e ElevUpdate) String() string {
	return "TID:" + strconv.Itoa(e.TransID) + ", RecvID:" +
		strconv.Itoa(e.RecvID) + ", State:" + e.Direction
}

type Order struct {
	TransID   int
	RecvID    int
	Floor     int
	Direction string
}

type Ack struct {
	TransID int
	RecvID  int
}

type Orders struct {
	TransID int
	RecvID  int
	Orders  *[]Order
}

/////   Sett inn flere datastructer her   /////
