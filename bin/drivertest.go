package main

import(
	"elevator/driver"
	"elevator"
	"fmt"
	//"strconv"
	"time"
)



func main(){
	// Initialize variables
	sOrder_ch := make(chan elevator.Order)
	rOrder_ch := make(chan elevator.Order)
	// Initialize hardware
	elevator.Init(sOrder_ch,rOrder_ch)
    driver.Set_direction(driver.DIRECTION_DOWN)
	
	    if !driver.Elevator_init() {
            fmt.Printf("Unable to initialize elevator hardware!\n")
            
    }
	

    for{
        // Change direction when the elevator reaches top/bottom floor
        /**
		if driver.Get_floor_sensor_signal() == driver.N_FLOORS - 1 {
            driver.Set_direction(driver.DIRECTION_DOWN)
        } else if driver.Get_floor_sensor_signal() == 0 {
            driver.Set_direction(driver.DIRECTION_UP)
        }
		*/
		time.Sleep(1*time.Millisecond)		
		//println("Floor: "+strconv.Itoa(driver.Get_floor_sensor_signal()))

    }
}