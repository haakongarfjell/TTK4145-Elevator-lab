package elev_struct

import (
	"elevator-project/config"
	"elevator-project/local_elevator/elev_io"
)

const (
	N_FLOORS  int = config.N_FLOORS
	N_BUTTONS     = config.N_BUTTONS
)

type State int

const (
	IDLE State = 0
	MOVE       = 1
	DOOR       = 2
)

type Orders [N_FLOORS][N_BUTTONS]bool

type Elevator struct {
	Requests  Orders
	Floor     int
	Dirn      elev_io.MotorDirection
	State     State
	ID        int
	StuckFlag bool
}

type DirnStatePair struct {
	Dirn  elev_io.MotorDirection
	State State
}

type Light struct {
	Floor  int
	Button elev_io.ButtonType
	Value  bool
}

func InitializeElevator(id int) Elevator {
	elev_io.SetDoorOpenLamp(false)
	return Elevator{
		Floor: 0,
		Dirn:  elev_io.MD_Stop,
		State: IDLE,
		ID:    id,
	}
}

func RemoveHallOrders(elev Elevator) Elevator {
	new_elevator := elev

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			new_elevator.Requests[floor][btn] = false
		}
	}
	return new_elevator
}

func GetCabCalls(elev Elevator) [N_FLOORS]bool {
	var cab_calls [N_FLOORS]bool

	for i := 0; i < N_FLOORS; i++ {
		cab_calls[i] = elev.Requests[i][2]
	}
	return cab_calls
}

func SetCabLights(elev Elevator) {
	for f := 0; f < N_FLOORS; f++ {
		elev_io.SetButtonLamp(elev_io.ButtonType(elev_io.BT_Cab), f, elev.Requests[f][elev_io.BT_Cab])
	}
}

func StateToString(state State) string {
	switch state {
	case IDLE:
		return "idle"
	case DOOR:
		return "doorOpen"
	case MOVE:
		return "moving"
	}
	return "INVALID STATE"
}

func MotorDirectionToString(dirn elev_io.MotorDirection) string {
	switch dirn {
	case elev_io.MD_Up:
		return "up"
	case elev_io.MD_Down:
		return "down"
	case elev_io.MD_Stop:
		return "stop"
	}
	return "INVALID DIRECTION"
}
