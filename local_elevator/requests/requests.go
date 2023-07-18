package requests

import (
	"elevator-project/local_elevator/elev_struct"
	"elevator-project/local_elevator/elev_io"
	"elevator-project/config"
)

const (
	N_FLOORS  int = config.N_FLOORS
	N_BUTTONS     = config.N_BUTTONS
)

func RequestsAbove(elev elev_struct.Elevator) bool {
	for f := elev.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elev.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsBelow(elev elev_struct.Elevator) bool {
	for f := 0; f < elev.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elev.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsHere(elev elev_struct.Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if elev.Requests[elev.Floor][btn] {
			return true
		}
	}
	return false
}

func RequestsChooseDirection(elev elev_struct.Elevator) elev_struct.DirnStatePair {
	switch elev.Dirn {
	case elev_io.MD_Up:
		if RequestsAbove(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Up, elev_struct.MOVE}
		} else if RequestsHere(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Down, elev_struct.DOOR}
		} else if RequestsBelow(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Down, elev_struct.MOVE}
		} else {
			return elev_struct.DirnStatePair{elev_io.MD_Stop, elev_struct.IDLE}
		}

	case elev_io.MD_Down:
		if RequestsBelow(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Down, elev_struct.MOVE}
		} else if RequestsHere(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Up, elev_struct.DOOR}
		} else if RequestsAbove(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Up, elev_struct.MOVE}
		} else {
			return elev_struct.DirnStatePair{elev_io.MD_Stop, elev_struct.IDLE}
		}

	case elev_io.MD_Stop:
		if RequestsHere(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Stop, elev_struct.DOOR}
		} else if RequestsAbove(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Up, elev_struct.MOVE}
		} else if RequestsBelow(elev) {
			return elev_struct.DirnStatePair{elev_io.MD_Down, elev_struct.MOVE}
		} else {
			return elev_struct.DirnStatePair{elev_io.MD_Stop, elev_struct.IDLE}
		}

	default:
		return elev_struct.DirnStatePair{elev_io.MD_Stop, elev_struct.IDLE}
	}
}

func RequestsShouldStop(elev elev_struct.Elevator) bool {
	switch elev.Dirn {
	case elev_io.MD_Down:
		return elev.Requests[elev.Floor][elev_io.BT_HallDown] ||
			elev.Requests[elev.Floor][elev_io.BT_Cab] ||
			!RequestsBelow(elev)

	case elev_io.MD_Up:
		return elev.Requests[elev.Floor][elev_io.BT_HallUp] ||
			elev.Requests[elev.Floor][elev_io.BT_Cab] ||
			!RequestsAbove(elev)

	case elev_io.MD_Stop:
		fallthrough

	default:
		return true
	}
}

func RequestsShouldClearImmediately(elev elev_struct.Elevator, btnFloor int, btnType elev_io.ButtonType) bool {
	return elev.Floor == btnFloor && ((elev.Dirn == elev_io.MD_Up && btnType == elev_io.BT_HallUp) ||
		(elev.Dirn == elev_io.MD_Down && btnType == elev_io.BT_HallDown) ||
		(elev.Dirn == elev_io.MD_Stop || btnType == elev_io.BT_Cab))
}

func RequestsClearAtCurrentFloor(elev elev_struct.Elevator, flag_clear chan<- elev_io.ButtonEvent) elev_struct.Elevator {
	elev.Requests[elev.Floor][elev_io.BT_Cab] = false

	switch elev.Dirn {
	case elev_io.MD_Up:
		if !RequestsAbove(elev) && !elev.Requests[elev.Floor][elev_io.BT_HallUp] {
			if elev.Requests[elev.Floor][elev_io.BT_HallDown] {
				flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallDown}
			}
			elev.Requests[elev.Floor][elev_io.BT_HallDown] = false
		}
		if elev.Requests[elev.Floor][elev_io.BT_HallUp] {
			flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallUp}
		}
		elev.Requests[elev.Floor][elev_io.BT_HallUp] = false

	case elev_io.MD_Down:
		if !RequestsBelow(elev) && !elev.Requests[elev.Floor][elev_io.BT_HallDown] {
			if elev.Requests[elev.Floor][elev_io.BT_HallUp] {
				flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallUp}
			}
			elev.Requests[elev.Floor][elev_io.BT_HallUp] = false
		}
		if elev.Requests[elev.Floor][elev_io.BT_HallDown] {
			flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallDown}
		}
		elev.Requests[elev.Floor][elev_io.BT_HallDown] = false

	case elev_io.MD_Stop:
		fallthrough
		
	default:
		if elev.Requests[elev.Floor][elev_io.BT_HallUp] {
			flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallUp}
		}
		if elev.Requests[elev.Floor][elev_io.BT_HallDown] {
			flag_clear <- elev_io.ButtonEvent{elev.Floor, elev_io.BT_HallDown}
		}
		elev.Requests[elev.Floor][elev_io.BT_HallUp] = false
		elev.Requests[elev.Floor][elev_io.BT_HallDown] = false
	}
	return elev
}