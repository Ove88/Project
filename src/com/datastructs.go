package com

import (
	"strconv"
)

type ElevData struct {
	TransID   int
	RecvID    int
	State     int
	Position  int
	Direction string
	//Destinations   []int
}

func (e ElevData) RemoteID() int {
	return e.RecvID
}
func (e ElevData) String() string {
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
