// elevator.go
package elevator

import (
	"elevator/driver"
	"time"
)
	
const (
	numberOfFloors  int = 4
	numberOfButtons int = 3
	secDoorOpen     int = 3
)

var (
	button_ch        chan ButtonPush
	elevPos_ch       chan Order
	doorOpen_ch      chan bool
	currentPosition  int
	currentDirection int
	doorOpen         bool
	stopped          bool
)

type Order struct {
	OriginID  int
	Internal  bool
	Floor     int
	Direction int
}

type ButtonPush struct {
	Floor  int
	Button int
}

type Position struct {
	LastPosition int
	Direction    int
}

func Init(sOrder_ch <-chan Order, rOrder_ch chan<- Order, pos_ch chan Position)bool {
	if !driver.Init(){
		return false
		
	}
	stopped = false
	button_ch = make(chan ButtonPush,10)
	elevPos_ch = make(chan Order,10)
	doorOpen_ch = make(chan bool,5)
	go elevatorPositionHandler(pos_ch)
	go floorLampHandler()
	go stopLampHandler(pos_ch)
	go buttonHandler()
	go orderGenerator(rOrder_ch)
	go orderHandler(sOrder_ch)
	go doorHandler()
	return true
}

func doorHandler() {
	for {
		<-doorOpen_ch
		doorOpen = true
		driver.Set_door_open_lamp(true)
		time.Sleep(3 * time.Second)
		driver.Set_door_open_lamp(false)
		doorOpen = false
	}
}

func orderHandler(sOrder_ch <-chan Order) {
	for {
		order := <-sOrder_ch
		for doorOpen {
			time.Sleep(10 * time.Millisecond)
		}
		if !stopped {
			if (order.Floor == currentPosition) && currentDirection != -1 {
				if currentDirection == 1 {
					driver.Set_direction(driver.DIRECTION_UP)
				} else {
					driver.Set_direction(driver.DIRECTION_DOWN)
				}
			} else if order.Floor < currentPosition {
				currentDirection = 1
				driver.Set_direction(driver.DIRECTION_DOWN)
			} else if order.Floor > currentPosition {
				driver.Set_direction(driver.DIRECTION_UP)
				currentDirection = 0
			}
			elevPos_ch <- order
		}
	}
}

func orderGenerator(rOrder_ch chan<- Order) {
	for {
		buttonPush := <-button_ch
		if buttonPush.Button == 2 { // Internal order
			if buttonPush.Floor > currentPosition {
				rOrder_ch <- Order{0, true, buttonPush.Floor, 0}
			} else if buttonPush.Floor < currentPosition {
				rOrder_ch <- Order{0, true, buttonPush.Floor, 1}
			} else if currentDirection == -1 {
				doorOpen_ch <- true
			}
		} else if buttonPush.Floor == currentPosition && currentDirection == -1 {
			if !doorOpen{
					doorOpen_ch <- true
			}
		} else {
			rOrder_ch <- Order{0, false, buttonPush.Floor, buttonPush.Button}
		}
	}
}

func elevatorPositionHandler(pos_ch chan Position) {
	var pos, lastPos int
	var arrived bool
	var order Order
	var initialization bool
	initialization = false
	arrived = false
	for {
		if initialization {
			if driver.Get_floor_sensor_signal() != -1 {
				driver.Set_direction(driver.DIRECTION_STOP)
				break
			}
		} else {
			initialization = true
			driver.Set_direction(driver.DIRECTION_UP)
		}
	}
	for {
		select {
		case order = <-elevPos_ch:
			arrived = false
			continue
		case <-time.After(10 * time.Millisecond):

			pos = driver.Get_floor_sensor_signal()
			if pos != -1 && lastPos != pos {
				currentPosition = pos
				pos_ch <- Position{pos, currentDirection}
			}
			if pos == order.Floor && !arrived {
				arrived = true

				time.Sleep(200 * time.Millisecond)
				driver.Set_direction(driver.DIRECTION_STOP)
				if order.Internal {
					driver.Set_button_indicator(2, pos, false)
				} else {
					driver.Set_button_indicator(order.Direction, pos, false)
				}
				currentDirection = -1
				if !doorOpen{
					doorOpen_ch <- true	
				}			
				pos_ch <- Position{pos, currentDirection}
			}
			lastPos = pos
		}
	}
}

func floorLampHandler() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		time.Sleep(1 * time.Millisecond)
	}
}

func SetButtonLamp(button, floor int, state bool) {
	driver.Set_button_indicator(button, floor, state)
}

func stopLampHandler(pos_ch chan Position) {
	flag := false
	for {
		stopBtn := driver.Get_stop_signal()
		if stopBtn && !flag {
			if !stopped {
				stopped = true
				driver.Set_stop_lamp(stopped)
				driver.Set_direction(driver.DIRECTION_STOP)
				pos_ch <- Position{-1, -1}
			} else {
				stopped = false
				driver.Set_stop_lamp(stopped)
				driver.Set_direction(currentDirection)
				pos_ch <- Position{currentPosition, currentDirection}
			}
			flag = true
		} else if !stopBtn && flag {
			flag = false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func buttonHandler() {
	currButtons := make([][]int, numberOfFloors)
	prevButtons := make([][]int, numberOfFloors)
	for i := range currButtons {
		currButtons[i] = make([]int, 3)
		prevButtons[i] = make([]int, 3)
	}

	for {
		for floor := 0; floor < numberOfFloors; floor++ {
			for btn := 0; btn < numberOfButtons; btn++ {
				prevButtons[floor][btn] = currButtons[floor][btn]
				currButtons[floor][btn] = driver.Get_button_signal(btn, floor)

				if currButtons[floor][btn] != prevButtons[floor][btn] &&
					currButtons[floor][btn] == 1 {
					button_ch <- ButtonPush{floor, btn}
				}
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

}