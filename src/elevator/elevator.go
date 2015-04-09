// elevator.go
package elevator

import (
	"elevator/driver"
	"time"
	//"strconv"
	
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
	OrderID int
	Internal bool
	Floor     int
	Direction string
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
	elevPos_ch=make(chan int)
	go setFloorLight()
	go setStopLamp()
	go readButtons()
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
			println("up")
	 		driver.Set_direction(driver.DIRECTION_UP)
		}else{
			driver.Set_direction(driver.DIRECTION_STOP)
		}
		elevPos_ch<-order.Floor
		
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
		floor:=<-elevPos_ch
		//println(strconv.Itoa(floor))
		for{
			time.Sleep(1 * time.Millisecond)
		pos = driver.Get_floor_sensor_signal()
		if pos!= -1 && lastPos!=pos{
		pos_ch<-Pos{pos,0}
		}
		//println(strconv.Itoa(pos))
		if pos == floor{
			driver.Set_direction(driver.DIRECTION_STOP)
			break
		}
		time.Sleep(1 * time.Millisecond)
		lastPos = pos
		}
	}
}

func setFloorLight() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		time.Sleep(1 * time.Millisecond)
	}
}

func SetButtonLight(light ButtonPush){
	driver.Elev_set_button_lamp(light.Button,light.Floor,true)
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
					//println("floor:"+strconv.Itoa(floor)+"btn:"+strconv.Itoa(btn))
				}	
			}
		}
		time.Sleep(1*time.Millisecond)
	}
	
}
