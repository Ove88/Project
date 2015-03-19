// elevator.go
package elevator

import (
	"elevator/driver"
	"time"
	"strconv"
	
)
const (N_FLOORS int = 4
N_BUTTONS int =3
)
var (
	stopFlag    bool
	stopFlag_sh bool
	stopLampOn  bool
	button_ch chan ButtonPush
	elevPos_ch chan int
)

type Order struct {
	Floor     int
	Direction string
}

type ButtonPush struct {
	Floor 	int
	Button 	int
}

// Styrer heisen. Leser knapper og setter posisjon

func Init(sOrder_ch <-chan Order, rOrder_ch chan<- Order) {
	button_ch = make(chan ButtonPush)
	go setFloorLight()
	go setStopLamp()
	go readButtons()
	go generateOrder(rOrder_ch)
}

func orderToExecute(sOrder_ch <-chan Order) {
	for{
		order:=<-sOrder_ch
		if order.Direction == "down"{
	 		driver.Set_direction(driver.DIRECTION_DOWN)
		}else if order.Direction == "up"{
	 		driver.Set_direction(driver.DIRECTION_UP)
		}
		elevPos_ch<-order.Floor
		
	}
}
func generateOrder(rOrder_ch chan<- Order) {
	for{
		buttonPush := <- button_ch
		println("Floor: "+ strconv.Itoa(buttonPush.Floor) + ", Button: "+strconv.Itoa(buttonPush.Button))
	}
}

func readElevatorPosition(){
	var pos int
	for{
		floor:=<-elevPos_ch
		driver.Get_floor_sensor_signal()
		time.Sleep(1 * time.Millisecond)
	}
}

func setFloorLight() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		time.Sleep(1 * time.Millisecond)
	}
}

func SetButtonLight(light ButtonPush){
	driver.Elev_set_button_lamp(light.Floor,light.Button,true)
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
		time.Sleep(1 * time.Millisecond)
	}
}

func readButtons(){
	currButtons := make([][]int,N_FLOORS)
    prevButtons :=  make([][]int,N_FLOORS)
	for i :=range currButtons{
		currButtons[i]=make([]int,3)
		prevButtons[i]=make([]int,3)
	}
	
	for{
		for floor := 0; floor < N_FLOORS; floor++{
			for  btn := 0;btn < N_BUTTONS; btn++{
				/*if btn == BUTTON_CALL_UP && floor == N_FLOORS-1 ||
				btn == BUTTON_CALL_DOWN && floor == 0{
					continue
				}*/
				
				prevButtons[floor][btn] = currButtons[floor][btn]
				currButtons[floor][btn] = driver.Get_button_signal(btn,floor)
				
				if currButtons[floor][btn] != prevButtons[floor][btn] && currButtons[floor][btn]==1{
					
					button_ch <- ButtonPush{floor,btn}
					SetButtonLight(ButtonPush{floor,btn})
				}
				
			}
		}
		time.Sleep(1*time.Millisecond)
	}
	
}
