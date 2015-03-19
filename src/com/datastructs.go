package com

import (
	"strconv"
)

type ElevData struct {
	TransactionID int
	ClientID      int
	State         int
	Position      int
	Direction     string
	//Destinations   []int
}

func (e ElevData) RemoteID() int {
	return e.ClientID
}
func (e ElevData) String() string {
	return "TID:" + strconv.Itoa(e.TransactionID) + ", ClientID:" +
		strconv.Itoa(e.ClientID) + ", State:" + e.Direction
}

type Order struct {
	ClientID  int
	Floor     int
	Direction string
}

/////   Sett inn flere datastructer her   /////

// type ElevOrder struct {
// 	ClientID     	 int
// 	Elev_destination int
//}
