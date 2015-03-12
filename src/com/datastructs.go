package com

import (
	"strconv"
)

type ElevData struct {
	Transaction_id int
	Client_id      int
	State          int
	Position       int
	Direction      string
	//Destinations   []int
}

func (e ElevData) RemoteID() int {
	return e.Client_id
}
func (e ElevData) String() string {
	return "TID:" + strconv.Itoa(e.Transaction_id) + ", ClientID:" +
		strconv.Itoa(e.Client_id) + ", State:" + e.Direction
}

/////   Sett inn flere datastructer her   /////

// type ElevOrder struct {
// 	Client_id     	 int
// 	Elev_destination int
//}
