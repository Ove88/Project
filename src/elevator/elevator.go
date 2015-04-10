// elevator.go
package elevator

import (
	"elevator/driver"
	"time"
	"strconv"
	
)
const (
	numberOfFloors 	int = 4
	numberOfButtons 	int = 3
	
)
var (
	stopFlag    bool
	stopFlag_sh bool
	stopLampOn  bool
	button_ch   chan ButtonPush
	elevPos_ch  chan Order
	doorOpen_ch chan bool
	currentPosition int
	currentDirection int
	doorOpen bool
)

type Order struct {
	Internal bool
	Floor     int
	Direction int
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
	driver.Set_direction(driver.DIRECTION_STOP)
	button_ch = make(chan ButtonPush)
	elevPos_ch=make(chan Order)
	doorOpen_ch = make(chan bool)
	go readElevatorPosition(pos_ch)
	go setFloorLamp()
	go setStopLamp()
	go buttonReader()
	go orderGenerator(rOrder_ch)
	go orderHandler(sOrder_ch)
	go doorHandler()
}

func doorHandler(){
	for{
	<-doorOpen_ch
	doorOpen = true
	driver.Set_door_open_lamp(true)
	time.Sleep(3*time.Second)
	driver.Set_door_open_lamp(false)
	doorOpen = false
	}
}

func orderHandler(sOrder_ch <-chan Order) {
	for{
		order:=<-sOrder_ch
		currentPosition = driver.Get_floor_sensor_signal()
		//println(strconv.Itoa(currentPosition))
		if doorOpen{
			time.Sleep(1*time.Millisecond)
			continue
		}else{
		if order.Floor < currentPosition{
			currentDirection = 1
			driver.Set_direction(driver.DIRECTION_DOWN)
		}else if order.Floor > currentPosition {
			driver.Set_direction(driver.DIRECTION_UP)
			currentDirection = 0
		}
		}
		elevPos_ch<-order
	}
}

func orderGenerator(rOrder_ch chan<- Order) {
	for{
		buttonPush := <- button_ch
		if buttonPush.Button == 2{
			println("ButtonPush.Floor: "+ strconv.Itoa(buttonPush.Floor) + ", CurrentPosition: "+strconv.Itoa(currentPosition))
			if buttonPush.Floor > currentPosition{
				println("er her 4")
				rOrder_ch<-Order{true,buttonPush.Floor,0}
			}else if buttonPush.Floor < currentPosition{
				println("er her 5")
				rOrder_ch<-Order{true,buttonPush.Floor,1}
			}else{
				println("er her 6")
				doorOpen_ch<-true
			}
		}else{
			rOrder_ch<-Order{false,buttonPush.Floor,buttonPush.Button}
		}
		//println("Floor: "+ strconv.Itoa(buttonPush.Floor) + ", Button: "+strconv.Itoa(buttonPush.Button))
	}
}

func readElevatorPosition(pos_ch chan Pos){
	var pos,lastPos int
	for{
		order:=<-elevPos_ch
		
		for{
			println("er her 1")
			pos = driver.Get_floor_sensor_signal()
			if pos!= -1 && lastPos!=pos{
				println("er her 2")
				currentPosition = pos
				println("currentPos:"+strconv.Itoa(currentPosition))
		      	pos_ch<-Pos{pos,currentDirection}
			}
			//println(strconv.Itoa(pos))
			if pos == order.Floor{
				println("er her 3")
				driver.Set_direction(driver.DIRECTION_STOP)
				if order.Internal{
					driver.Set_button_indicator(2,pos,false)
				}else{
					driver.Set_button_indicator(order.Direction,pos,false)					
				}
				currentDirection = -1
				doorOpen_ch<-true
				pos_ch<-Pos{pos,currentDirection}
				break
			}
			time.Sleep(1 * time.Millisecond)
			lastPos = pos
		}
	}
}

func setFloorLamp() {
	for {
		driver.Set_floor_indicator(driver.Get_floor_sensor_signal())
		//println(strconv.Itoa(currentPosition))
		time.Sleep(1 * time.Millisecond)
	}
}

func SetButtonLamp(button, floor int){
		driver.Set_button_indicator(button,floor,true)
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
				driver.Set_direction(driver.DIRECTION_STOP)
			} else {
				stoplamp = false
				driver.Set_stop_lamp(stoplamp)
				driver.Set_direction(currentDirection)
			}
			flag = true
		}else if !stopBtn && flag{
			flag = false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func buttonReader(){
	currButtons := make([][]int,numberOfFloors)
    prevButtons :=  make([][]int,numberOfFloors)
	for i :=range currButtons{
		currButtons[i]=make([]int,3)
		prevButtons[i]=make([]int,3)
	}
	
	for{
		for floor := 0; floor < numberOfFloors; floor++{
			for  btn := 0;btn < numberOfButtons; btn++{
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
