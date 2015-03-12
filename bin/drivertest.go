package main

import "driver"
import "fmt"


func main(){
	// Initialize hardware
	    if !driver.Elevator_init() {
            fmt.Printf("Unable to initialize elevator hardware!\n")
            
    }

    fmt.Printf("Press STOP button to stop elevator and exit program.\n")
    driver.Set_direction(driver.DIRECTION_DOWN)

    for{
        // Change direction when we reach top/bottom floor
        if driver.Get_floor_sensor_signal() == driver.N_FLOORS - 1 {
            driver.Set_direction(driver.DIRECTION_DOWN)
        } else if driver.Get_floor_sensor_signal() == 0 {
            driver.Set_direction(driver.DIRECTION_UP)
        }

        // Stop elevator and exit program if the stop button is pressed
        if driver.Get_stop_signal() {
            driver.Set_direction(driver.DIRECTION_STOP)
            break
        }
    }


}

