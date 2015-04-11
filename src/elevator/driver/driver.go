package driver


// Number of signals and lamps on a per-floor basis (excl sensor)
const N_BUTTONS = 3
const N_FLOORS = 4



type Elev_btn_type int

/*
const (
	BTN_UP Elev_btn_type = iota
	BTN_DOWN
	BTN_COMMAND
)
*/

const (
	DIRECTION_UP int = 0
	DIRECTION_STOP int = -1
	DIRECTION_DOWN int = 1
)

var lamp_channel_matrix = [N_FLOORS][N_BUTTONS]int {
    {LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
    {LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
    {LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
    {LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}


var button_channel_matrix = [N_FLOORS][N_BUTTONS]int {
    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func Elevator_init() bool{

    // Init hardware
    if !IO_init(){
        return false
	}

    // Zero all floor button lamps
    for i := 0; i < N_FLOORS; i++ {
        if i != 0{
           Set_button_indicator(0, i, false)
	}

        if i != N_FLOORS - 1 {
           Set_button_indicator(1, i, false)
	}
       Set_button_indicator(2, i, false)
    }

    // Clear stop lamp, door open lamp, and set floor indicator to ground floor.
    Set_stop_lamp(false)
    Set_door_open_lamp(false)
    Set_floor_indicator(0)

    // Return success.
    return true
}

func Set_direction(direction int) {
	switch direction{
	case DIRECTION_UP:
		Clear_bit(MOTORDIR)
		Write_analog(MOTOR,2400)
	case DIRECTION_DOWN:
		Set_bit(MOTORDIR)
		Write_analog(MOTOR,2400)
	case DIRECTION_STOP:
		Write_analog(MOTOR,0)
	}
}

func Set_door_open_lamp(lampOpen bool) {
    if lampOpen{
        Set_bit(LIGHT_DOOR_OPEN)
    }else{
        Clear_bit(LIGHT_DOOR_OPEN)
    }
}

/*
func get_obstruction_signal()int {
    return Read_bit(OBSTRUCTION)
}*/

func Get_stop_signal()bool {
    return Read_bit(STOP)
}

func Set_stop_lamp(lampStop bool) {
    if lampStop{
        Set_bit(LIGHT_STOP)
    }else{
        Clear_bit(LIGHT_STOP)
    }
}

func Get_floor_sensor_signal()int {
    if (Read_bit(SENSOR_FLOOR1)){
        return 0
    }else if (Read_bit(SENSOR_FLOOR2)){
        return 1
    }else if (Read_bit(SENSOR_FLOOR3)){
        return 2
    }else if (Read_bit(SENSOR_FLOOR4)){
        return 3
    }else{
        return -1
    }
}

func Set_floor_indicator(floor int) {
	switch floor{
	case 0:
		Clear_bit(LIGHT_FLOOR_IND1)
		Clear_bit(LIGHT_FLOOR_IND2)
	case 1:
		Clear_bit(LIGHT_FLOOR_IND1)
		Set_bit(LIGHT_FLOOR_IND2)
	case 2:
		Set_bit(LIGHT_FLOOR_IND1)
		Clear_bit(LIGHT_FLOOR_IND2)
	case 3:
		Set_bit(LIGHT_FLOOR_IND1)
		Set_bit(LIGHT_FLOOR_IND2)
	}
}


func Get_button_signal(button int, floor int)int{ 
    if Read_bit(button_channel_matrix[floor][button]){
        return 1
	}else{
        return 0
	}
}

func Set_button_indicator( button int, floor int, value bool) {
    if value{
        Set_bit(lamp_channel_matrix[floor][button])
    }else{
        Clear_bit(lamp_channel_matrix[floor][button])
	}
}
