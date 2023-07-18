package elev_state_machine

import (
	"elevator-project/local_elevator/elev_struct"
	"elevator-project/local_elevator/requests"
	"elevator-project/local_elevator/elev_io"
	"elevator-project/config"
	"time"
)

const (
	T_DOOR  = config.T_DOOR
	T_STUCK = config.T_STUCK
)

func OnInitBetweenFloors(elev elev_struct.Elevator) elev_struct.Elevator {
	operating_elevator := elev
	elev_io.SetMotorDirection(elev_io.MD_Down)
	operating_elevator.Dirn = elev_io.MD_Down
	operating_elevator.State = elev_struct.MOVE
	return operating_elevator
}

func OnRequestButtonPress(
	elev elev_struct.Elevator, 
	btn_floor int, 
	btn_type elev_io.ButtonType, 
	door_timer *time.Timer, 
	stuck_timer *time.Timer, 
	flag_clear_ch chan<- elev_io.ButtonEvent) elev_struct.Elevator {

	operating_elevator := elev

	switch operating_elevator.State {
	case elev_struct.DOOR:
		if requests.RequestsShouldClearImmediately(operating_elevator, btn_floor, btn_type) {
			flag_clear_ch <- elev_io.ButtonEvent{btn_floor, btn_type}
			door_timer.Reset(T_DOOR * time.Second)
			stuck_timer.Reset(T_STUCK * time.Second)
		} else {
			operating_elevator.Requests[btn_floor][btn_type] = true
		}

	case elev_struct.MOVE:
		operating_elevator.Requests[btn_floor][btn_type] = true

	case elev_struct.IDLE:
		operating_elevator.Requests[btn_floor][btn_type] = true
		nxt_action := requests.RequestsChooseDirection(operating_elevator)
		operating_elevator.Dirn = nxt_action.Dirn
		operating_elevator.State = nxt_action.State
		switch nxt_action.State {
		case elev_struct.DOOR:
			elev_io.SetDoorOpenLamp(true)
			door_timer.Reset(T_DOOR * time.Second)
			stuck_timer.Reset(T_STUCK * time.Second)
			operating_elevator = requests.RequestsClearAtCurrentFloor(operating_elevator, flag_clear_ch)

		case elev_struct.MOVE:
			elev_io.SetMotorDirection(operating_elevator.Dirn)
			stuck_timer.Reset(T_STUCK * time.Second)

		case elev_struct.IDLE:
		}
	}
	elev_struct.SetCabLights(operating_elevator)
	return operating_elevator
}

func OnFloorArrival(
	elev elev_struct.Elevator, 
	new_floor int, 
	door_timer *time.Timer, 
	flag_clear_ch chan<- elev_io.ButtonEvent) elev_struct.Elevator {

	operating_elevator := elev
	operating_elevator.Floor = new_floor
	elev_io.SetFloorIndicator(operating_elevator.Floor)

	switch operating_elevator.State {
	case elev_struct.MOVE:
		if requests.RequestsShouldStop(operating_elevator) {
			elev_io.SetMotorDirection(elev_io.MD_Stop)
			elev_io.SetDoorOpenLamp(true)
			operating_elevator = requests.RequestsClearAtCurrentFloor(operating_elevator, flag_clear_ch)
			door_timer.Reset(T_DOOR * time.Second)
			elev_struct.SetCabLights(operating_elevator)
			operating_elevator.State = elev_struct.DOOR
		}

	default:
		break
	}
	return operating_elevator
}

func OnDoorTimeout(
	elev elev_struct.Elevator, 
	door_timer *time.Timer, 
	flag_clear_ch chan<- elev_io.ButtonEvent) elev_struct.Elevator {

	operating_elevator := elev

	switch operating_elevator.State {
	case elev_struct.DOOR:
		nxt_action := requests.RequestsChooseDirection(operating_elevator)
		operating_elevator.Dirn = nxt_action.Dirn
		operating_elevator.State = nxt_action.State

		switch operating_elevator.State {
		case elev_struct.DOOR:
			door_timer.Reset(T_DOOR * time.Second)
			operating_elevator = requests.RequestsClearAtCurrentFloor(operating_elevator, flag_clear_ch)
			elev_struct.SetCabLights(operating_elevator)

		case elev_struct.MOVE:
			fallthrough

		case elev_struct.IDLE:
			elev_io.SetDoorOpenLamp(false)
			elev_io.SetMotorDirection(operating_elevator.Dirn)
			elev_struct.SetCabLights(operating_elevator)
		}

	default:
		break
	}
	return operating_elevator
}

func OnObstruction(elev elev_struct.Elevator, door_timer *time.Timer) {
	operating_elevator := elev
	
	switch operating_elevator.State {
	case elev_struct.DOOR:
		door_timer.Reset(T_DOOR * time.Second)

	default:
		break
	}
}
