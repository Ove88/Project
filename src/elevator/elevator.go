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
	elevPos_ch chan Order
)

type Order struct {
	OrderID int
	Internal bool
	Floor     int
	Direction string
}

type Light struct{
	State bool
	Floor int
	Button int
}
type ButtonPush struct {
	Floor 	int
	Button 	int
}

type Pos struct{
	LastPos int
	Direction int
}
// Styrer heisen. Leser knapper og setter posisjon

func Init(sOrder_ch <-chan Order, rOrder_ch chan<- Order, pos_ch chan Pos) {
	button_ch = make(chan ButtonPush)
	elevPos_ch=make(chan Order)
	go setFloorLamp()
	go setStopLamp()
	go buttonReader()
	go generateOrder(rOrder_ch)
	go orderToExecute(sOrder_ch)
	go readElevatorPosition(pos_ch)
}

func orderToExecute(sOrder_ch <-chan Order) {
	for{
		order:=<-sOrder_ch
		if order.Direction == "down"{
	 		driver.Set_direction(driver.DIRECTION_DOWN)
		}else if order.Direction == "up"{
	 		driver.Set_direction(driver.DIRECTION_UP)
		}else{
			driver.Set_direction(driver.DIRECTION_STOP)
		}
		elevPos_ch<-order
		
	}
}

func generateOrder(rOrder_ch chan<- Order) {
	for{
		buttonPush := <- button_ch
		if buttonPush.Button == 2{
		rOrder_ch<-Order{0,true,buttonPush.Floor,""}
		}else if buttonPush.Button ==0{
			
			rOrder_ch<-Order{0,false,buttonPush.Floor,"up"}
		}else{
			rOrder_ch<-Order{0,false,buttonPush.Floor,"down"}
		}
		//println("Floor: "+ strconv.Itoa(buttonPush.Floor) + ", Button: "+strconv.Itoa(buttonPush.Button))
	}
}

func readElevatorPosition(pos_ch chan Pos){
	var pos,lastPos int
	for{
		order:=<-elevPos_ch
		//println(strconv.Itoa(floor))
		for{
			time.Sleep(1 * time.Millisecond)
		pos = driver.Get_floor_sensor_signal()
		if pos!= -1 && lastPos!=pos{
		pos_ch<-Pos{pos,0}
		}
		//println(strconv.Itoa(pos))
		if pos == order.Floor{
			driver.Set_direction(driver.DIRECTION_STOP)
			if order.Direction == "up"{
				driver.
				//SetButtonLight(ButtonPush{order.Floor,0},false)
			}else{
				//SetButtonLight(ButtonPush{order.Floor,1},false)
			}
			driver.Set_door_open_lamp(true)
			break
		}else{
			driver.Set_door_open_lamp(false)
		}
		time.Sleep(1 * time.Millisecond)
		lastPos = pos
		}
	}
}

func setFloorLamp() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		time.Sleep(1 * time.Millisecond)
	}
}

func SetButtonLamp(button, floor int){
		driver.Elev_set_button_lamp(button,floor,true)
}

func setStopLamp() {
	var stoplamp bool
	var flag bool
	for {
		stopBtn := driver.Get_stop_signal()
		if stopBtn && !flag {
			if !stoplamp {
				stoplamp = true
				driver.Set_stop_lamp(stoplamp)
			} else {
				stoplamp = false
				driver.Set_stop_lamp(stoplamp)
			}
			flag = true
		}else if !stopBtn && flag{
			flag = false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func buttonReader(){
	currButtons := make([][]int,N_FLOORS)
    prevButtons :=  make([][]int,N_FLOORS)
	for i :=range currButtons{
		currButtons[i]=make([]int,3)
		prevButtons[i]=make([]int,3)
	}
	
	for{
		for floor := 0; floor < N_FLOORS; floor++{
			for  btn := 0;btn < N_BUTTONS; btn++{
				prevButtons[floor][btn] = currButtons[floor][btn]
				currButtons[floor][btn] = driver.Get_button_signal(btn,floor)
				
				if currButtons[floor][btn] != prevButtons[floor][btn] && 
				currButtons[floor][btn]==1{			
					button_ch <- ButtonPush{floor,btn}
				}
			}
		}
		time.Sleep(1*time.Millisecond)
	}
	
}
