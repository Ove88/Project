// elevator.go
package elevator

import (
	"elevator/driver"
	"time"
)

var (
	stopFlag    bool
	stopFlag_sh bool
	stopLampOn  bool
)

type Order struct {
	Floor     int
	Direction string
}

// Styrer heisen. Leser knapper og setter posisjon

func Init(setFloor_ch <-chan int, getFloor_ch chan<- int, button_ch chan<- int) {
	go setFloorLight()
	go setStopLamp()
}

func setFloor(setFloor_ch <-chan int) {

}
func getFloor(getFloor_ch chan<- int) {

}

func readButtons(button_ch chan<- int) {

}

func setFloorLight() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		time.Sleep(1 * time.Millisecond)
	}
}

func setStopLamp() {
	for {
		stopFlag = driver.Get_stop_signal()
		if stopFlag != stopFlag_sh {
			if !stopLampOn {
				driver.Set_stop_lamp(stopFlag)
				stopLampOn = true
			} else {
				driver.Set_stop_lamp(stopFlag)
				stopLampOn = false
			}
		}
		stopFlag_sh = stopFlag

		/**
				if time.Since(t)> 350000000 {
					if driver.Get_stop_signal() && !stopFlag {
						driver.Set_stop_lamp(true)
						stopFlag = true
						t = time.Now()
		        		}else if driver.Get_stop_signal() && stopFlag{
						driver.Set_stop_lamp(false)
						stopFlag = false
						t = time.Now()
					}
					time.Sleep(1*time.Millisecond)
				}
		*/
	}
}
