package driver
/*
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"

func IO_init() bool {
	return bool(int(C.io_init()) != 1)
}

func Set_bit(bit int) {
    C.io_set_bit(C.int(bit))
}

func Clear_bit(bit int) {
    C.io_clear_bit(C.int(bit))
}

func Write_analog(bit , value int) {
    C.io_write_analog(C.int(bit),C.int(value))
}

func Read_bit(bit int) bool{
	return bool(int(C.io_read_bit(C.int(bit))) != 0)
}

func Read_analog(bit int) int {
	return int(C.io_read_analog(C.int(bit)))
}

