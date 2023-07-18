package single_elevator

import (
	"elevator-project/config"
	"elevator-project/local_elevator/elev_io"
	"elevator-project/local_elevator/elev_state_machine"
	"elevator-project/local_elevator/elev_struct"
	"time"
)

const (
	T_DOOR        = config.T_DOOR
	T_STUCK       = config.T_STUCK
	BETWEEN_FLOOR = config.BETWEEN_FLOOR
)

func SingleElevatorRun(
	Id int,
	drv_floors_ch chan int,
	drv_obstr_ch chan bool,
	drv_buttons_ch chan elev_io.ButtonEvent,
	clear_local_hall_ch chan bool,
	clear_order_ch chan<- elev_io.ButtonEvent,
	elev_out_ch chan<- elev_struct.Elevator,
	assigned_orders_ch chan elev_io.ButtonEvent) {

	door_timer := time.NewTimer(T_DOOR * time.Second)
	stuck_timer := time.NewTimer(T_STUCK * time.Second)

	operating_elevator := elev_struct.InitializeElevator(Id)
	operating_elevator.Floor = elev_io.GetFloor()

	for {
		select {
		case <-clear_local_hall_ch:
			operating_elevator = elev_struct.RemoveHallOrders(operating_elevator)

		case btn_event := <-drv_buttons_ch:
			elev_out := operating_elevator
			elev_out.Requests[btn_event.Floor][btn_event.Button] = true
			elev_out_ch <- elev_out
			if btn_event.Button == elev_io.BT_Cab {
				assigned_orders_ch <- elev_io.ButtonEvent{btn_event.Floor, btn_event.Button}
			}

		case new_floor := <-drv_floors_ch:
			operating_elevator = elev_state_machine.OnFloorArrival(operating_elevator, new_floor, door_timer, clear_order_ch)
			stuck_timer.Reset(T_STUCK * time.Second)
			if operating_elevator.StuckFlag {
				operating_elevator.StuckFlag = false
			}

		case obstruction := <-drv_obstr_ch:
			if obstruction {
				elev_state_machine.OnObstruction(operating_elevator, door_timer)
			}

		case <-stuck_timer.C:
			if operating_elevator.State == elev_struct.MOVE || operating_elevator.State == elev_struct.DOOR {
				operating_elevator.StuckFlag = true
				operating_elevator = elev_struct.RemoveHallOrders(operating_elevator)
			}

		case <-door_timer.C:
			operating_elevator = elev_state_machine.OnDoorTimeout(operating_elevator, door_timer, clear_order_ch)
			stuck_timer.Reset(T_STUCK * time.Second)
			if operating_elevator.StuckFlag {
				operating_elevator.StuckFlag = false
			}

		case btn_event := <-assigned_orders_ch:
			operating_elevator = elev_state_machine.OnRequestButtonPress(operating_elevator, btn_event.Floor, btn_event.Button, door_timer, stuck_timer, clear_order_ch)

		default:
			elev_out_ch <- operating_elevator
		}
	}
}
