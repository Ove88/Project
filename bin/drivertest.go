package main

import(
	"elevator/driver"
	"elevator"
	"fmt"
	"strconv"
	//"time"
)



func main(){
	// Initialize variables
	sOrder_ch := make(chan elevator.Order,2)
	rOrder_ch := make(chan elevator.Order,2)
	pos_ch:=make(chan elevator.Pos)
	
	// Initialize hardware
	elevator.Init(sOrder_ch,rOrder_ch,pos_ch)
    driver.Set_direction(driver.DIRECTION_DOWN)
	
	    if !driver.Elevator_init() {
            fmt.Printf("Unable to initialize elevator hardware!\n")
    }
	go func(){
		for{
		order:=<-rOrder_ch
		println("order:"+strconv.Itoa(order.Floor))
		if order.Internal{
			elevator.SetButtonLamp(2,order.Floor)
		}else{
			elevator.SetButtonLamp(order.Direction,order.Floor)
		}	
		sOrder_ch<-order
		}
	}()
	//sOrder_ch<-elevator.Order{false,1,0}
    for{
		pos:=<-pos_ch
		println("pos:"+strconv.Itoa(pos.LastPos))
		//order:=<-rOrder_ch
		//println("order:"+strconv.Itoa(order.Floor))
		//if order.Internal{
		//	elevator.SetButtonLamp(2,order.Floor)
		//}else{
		//	elevator.SetButtonLamp(order.Direction,order.Floor)
		//}
		
		//sOrder_ch<-order
		//pos:=<-pos_ch
		//println("pos:"+strconv.Itoa(pos.LastPos))
		//println(order.Direction)
        // Change direction when the elevator reaches top/bottom floor
        /**
		if driver.Get_floor_sensor_signal() == driver.N_FLOORS - 1 {
            driver.Set_direction(driver.DIRECTION_DOWN)
        } else if driver.Get_floor_sensor_signal() == 0 {
            driver.Set_direction(driver.DIRECTION_UP)
        }
		*/
		//time.Sleep(1*time.Millisecond)		
		//println("Floor: "+strconv.Itoa(driver.Get_floor_sensor_signal()))

    }
}
