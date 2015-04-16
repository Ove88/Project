package main

import (
	"com"
	"elevator"
)

func main() {
	wait := make(chan bool)
	prc := com.NewHeaderProtocol{1000}
	pr := prc.NewProtocol()
	data := elevator.Order{true, 4, 0}
	data2 := com.Order{true, 2, 5}
	data3 := com.Order{true, 1, 3}
	orders := com.Orders{[]*com.Order{&data, &data2, &data3}}
	message := com.Header{100, 1, 2, orders}
	raw := pr.Encode(message)
	recv, _ := pr.Decode(raw)
	f := recv.(com.Header)
	println(f.GetType())
	o := f.Data.(com.Orders)
	println(o.Orders[0].String())

	// switch recv.Data.(type) {
	// case com.Order:
	// 	println("rOrder")
	// default:
	// 	println("Default")
	// }
	<-wait
}
