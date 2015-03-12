package driver


// Number of signals and lamps on a per-floor basis (excl sensor)
const N_BUTTONS = 3
const N_FLOORS = 4



/*
type Elev_btn_type int

const (
	BTN_UP Elev_btn_type = iota
	BTN_DOWN
	BTN_COMMAND
)
*/

const (
	DIRECTION_UP int =1
	DIRECTION_STOP int = 0
	DIRECTION_DOWN int = -1
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
           // Set_button_lamp(BUTTON_CALL_DOWN, i, 0)
	}

        if i != N_FLOORS - 1 {
           // Set_button_lamp(BUTTON_CALL_UP, i, 0)
	}
       // Set_button_lamp(BUTTON_COMMAND, i, 0)
    }

    // Clear stop lamp, door open lamp, and set floor indicator to ground floor.
    set_stop_lamp(false)
    set_door_open_lamp(false)
    set_floor_indicator(0)

    // Return success.
    return true
}

func Set_direction(direction int) {
	switch direction{
	case DIRECTION_UP:
		Clear_bit(MOTORDIR)
		Write_analog(MOTOR,2800)
	case DIRECTION_DOWN:
		Set_bit(MOTORDIR)
		Write_analog(MOTOR,2800)
	case DIRECTION_STOP:
		Write_analog(MOTOR,0)
	}
}

func set_door_open_lamp(lampOpen bool) {
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

func set_stop_lamp(lampStop bool) {
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

func set_floor_indicator(floor int) {
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


/*
func get_button_signal(elev_button_type_t button, int floor) {
    assert(floor >= 0);
    assert(floor < N_FLOORS);
    assert(!(button == BUTTON_CALL_UP && floor == N_FLOORS - 1));
    assert(!(button == BUTTON_CALL_DOWN && floor == 0));
    assert(button == BUTTON_CALL_UP || button == BUTTON_CALL_DOWN || button == BUTTON_COMMAND);

    if (io_read_bit(button_channel_matrix[floor][button]))
        return 1;
    else
        return 0;
}


func elev_set_button_lamp(elev_button_type_t button, int floor, int value) {
    assert(floor >= 0);
    assert(floor < N_FLOORS);
    assert(!(button == BUTTON_CALL_UP && floor == N_FLOORS - 1));
    assert(!(button == BUTTON_CALL_DOWN && floor == 0));
    assert(button == BUTTON_CALL_UP || button == BUTTON_CALL_DOWN || button == BUTTON_COMMAND);

    if (value)
        io_set_bit(lamp_channel_matrix[floor][button]);
    else
        io_clear_bit(lamp_channel_matrix[floor][button]);
}
*/
