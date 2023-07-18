package orders

import (
	"elevator-project/local_elevator/backup"
	"elevator-project/config"
	"elevator-project/distributor"
	"elevator-project/local_elevator/elev_io"
	"elevator-project/local_elevator/elev_struct"
	"elevator-project/network/peers"
	"fmt"
	"strconv"
	"time"
)

const (
	N_ELEVATORS int           = config.N_ELEVATORS
	N_FLOORS    int           = config.N_FLOORS
	N_BUTTONS   int           = config.N_BUTTONS
	T_SLEEP     time.Duration = config.T_SLEEP
)

type HallState int

const (
	NONE      HallState = 0
	NEW                 = 1
	CONFIRMED           = 2
	ASSIGNED            = 3
	COMPLETE            = 4
)

type ElevatorStates [N_ELEVATORS]elev_struct.Elevator

type HallOrders [N_FLOORS][N_BUTTONS - 1]HallState

type NetworkMessage struct {
	Elevator   elev_struct.Elevator
	HallOrders [N_FLOORS][N_BUTTONS - 1]HallState
}

func InitializeHallOrders() [N_FLOORS][N_BUTTONS - 1]HallState {
	var hall_orders [N_FLOORS][N_BUTTONS - 1]HallState

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			elev_io.SetButtonLamp(elev_io.ButtonType(btn), floor, false)
			hall_orders[floor][btn] = NONE
		}
	}
	return hall_orders
}

func InitializeAllHallOrders() [N_ELEVATORS]HallOrders {
	var all_hall_orders [N_ELEVATORS]HallOrders

	for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
		all_hall_orders[elev_id] = InitializeHallOrders()
	}
	return all_hall_orders
}

func InitializeAllElevators() ElevatorStates {
	var elevators ElevatorStates

	for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
		elevators[elev_id] = elev_struct.InitializeElevator(elev_id)
	}
	return elevators
}

func InitializeAllStuckFlags() [N_ELEVATORS]bool {
	var stuck_flags [N_ELEVATORS]bool

	for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
		stuck_flags[elev_id] = false
	}
	return stuck_flags
}

func FormatToDistributor(
	all_elevs ElevatorStates,
	id_int int,
	confirmed_hall_orders [N_ELEVATORS]HallOrders,
	available_elevators [N_ELEVATORS]bool,
	stuck_elevators [N_ELEVATORS]bool) distributor.HRAInput {

	var hall_orders [N_FLOORS][N_BUTTONS - 1]bool
	distr_elev_states := make(map[string]distributor.HRAElevState)

	for index, elev := range all_elevs {
		var cab_requests [N_FLOORS]bool

		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS-1; btn++ {
				if confirmed_hall_orders[id_int][floor][btn] == CONFIRMED {
					hall_orders[floor][btn] = true
				}
			}
			cab_requests[floor] = elev.Requests[floor][elev_io.BT_Cab]
		}

		distr_elev_states[strconv.Itoa(index)] = distributor.HRAElevState{
			Behavior:    elev_struct.StateToString(elev.State),
			Floor:       elev.Floor,
			Direction:   elev_struct.MotorDirectionToString(elev.Dirn),
			CabRequests: cab_requests,
		}
	}

	for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
		if !available_elevators[elev_id] || stuck_elevators[elev_id] {
			delete(distr_elev_states, strconv.Itoa(elev_id))
		}
	}
	return distributor.HRAInput{hall_orders, distr_elev_states}
}

func ResetHallOrders(
	Id int,
	all_hall_orders *[N_ELEVATORS]HallOrders,
	available_elevators *[N_ELEVATORS]bool,
	reset_orders_ch chan<- [N_ELEVATORS]HallOrders,
	hall_light_ch chan<- elev_struct.Light) {

	for {
		time.Sleep(T_SLEEP * time.Millisecond)
		new_all_hall_orders := *all_hall_orders
		current_available_elevators := *available_elevators

		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS-1; btn++ {

				should_reset := true

				for index, hall_orders := range new_all_hall_orders {
					if current_available_elevators[index] {
						switch hall_orders[floor][btn] {
						case COMPLETE:
							break
						default:
							should_reset = false
						}
					}
				}

				if should_reset {
					light_off := elev_struct.Light{floor, elev_io.ButtonType(btn), false}
					hall_light_ch <- light_off
					for index, _ := range new_all_hall_orders {
						if current_available_elevators[index] {
							new_all_hall_orders[index][floor][btn] = NONE
						}
					}
					reset_orders_ch <- new_all_hall_orders
				}
			}
		}
	}
}

func ConfirmHallOrders(
	Id int,
	all_hall_orders *[N_ELEVATORS]HallOrders,
	available_elevators *[N_ELEVATORS]bool,
	confirmed_orders_ch chan<- [N_ELEVATORS]HallOrders,
	hall_light_ch chan<- elev_struct.Light) {

	for {
		time.Sleep(T_SLEEP * time.Millisecond)
		new_all_hall_orders := *all_hall_orders
		current_available_elevators := *available_elevators

		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS-1; btn++ {

				should_confirm := true

				if current_available_elevators[Id] {
					for index, hall_orders := range new_all_hall_orders {
						if current_available_elevators[index] {
							switch hall_orders[floor][btn] {
							case NEW:
								break
							default:
								should_confirm = false
							}
						}
					}
				} else {
					should_confirm = false
				}

				if should_confirm {
					light_on := elev_struct.Light{floor, elev_io.ButtonType(btn), true}
					hall_light_ch <- light_on
					for index, _ := range new_all_hall_orders {
						if current_available_elevators[index] {
							new_all_hall_orders[index][floor][btn] = CONFIRMED
						}
					}
					confirmed_orders_ch <- new_all_hall_orders
				}
			}
		}
	}
}

func SetAvailableElevator(id_int int, state bool) [N_ELEVATORS]bool {
	var available_elevators [N_ELEVATORS]bool
	available_elevators[id_int] = state
	return available_elevators
}

func SetOrdersAssigned(Id int, all_hall_orders [N_ELEVATORS]HallOrders) [N_ELEVATORS]HallOrders {
	new_all_hall_orders := all_hall_orders

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			if new_all_hall_orders[Id][floor][btn] == CONFIRMED {
				new_all_hall_orders[0][floor][btn] = ASSIGNED
				new_all_hall_orders[1][floor][btn] = ASSIGNED
				new_all_hall_orders[2][floor][btn] = ASSIGNED
			}
		}
	}
	return new_all_hall_orders

}

func UpdateHallOrders(
	id_int int,
	elev elev_struct.Elevator,
	all_hall_orders [N_ELEVATORS]HallOrders) [N_ELEVATORS]HallOrders {

	new_all_hall_orders := all_hall_orders
	hall_orders := all_hall_orders[id_int]

	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS-1; button++ {
			if elev.Requests[floor][button] && (hall_orders[floor][button] == NONE) {
				hall_orders[floor][button] = NEW
			}
		}
	}
	new_all_hall_orders[id_int] = hall_orders
	return new_all_hall_orders
}

func UpdateFromDistributor(
	output map[string][][config.N_BUTTONS - 1]bool,
	all_elevs ElevatorStates,
	available_elevators [N_ELEVATORS]bool,
	stuck_elevators [N_ELEVATORS]bool) ElevatorStates {

	new_elevators := all_elevs

	for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS-1; btn++ {
				if available_elevators[elev_id] && !stuck_elevators[elev_id] {
					new_elevators[elev_id].Requests[floor][btn] = output[strconv.Itoa(elev_id)][floor][btn]
				}
			}
		}
	}
	return new_elevators
}

// A state machine based on an elevator's own all_hall_orders which holds cyclic counters. 
// The state machine receives all_hall_orders from other elevators and updates the cyclic counters accordingly.
func UpdateHallOrdersFromNetwork(
	Id int,
	msg NetworkMessage,
	all_hall_orders [N_ELEVATORS]HallOrders) [N_ELEVATORS]HallOrders {

	msg_Id := msg.Elevator.ID
	msg_hall_orders := msg.HallOrders
	new_all_hall_orders := all_hall_orders

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {

			switch msg_hall_orders[floor][btn] {
			case NONE:
				switch new_all_hall_orders[Id][floor][btn] {
				case COMPLETE:
					for elev_id := 0; elev_id < N_ELEVATORS; elev_id++ {
						new_all_hall_orders[elev_id][floor][btn] = COMPLETE
					}
				default:
					break
				}
				break

			case NEW:
				switch new_all_hall_orders[Id][floor][btn] {
				case NONE:
					new_all_hall_orders[Id][floor][btn] = NEW
				default:
					break
				}
				switch new_all_hall_orders[msg_Id][floor][btn] {
				case NONE:
					new_all_hall_orders[msg_Id][floor][btn] = NEW
				default:
					break
				}

			case CONFIRMED:
				switch new_all_hall_orders[Id][floor][btn] {
				case NONE:
					new_all_hall_orders[Id][floor][btn] = CONFIRMED
				case NEW:
					new_all_hall_orders[Id][floor][btn] = CONFIRMED
				default:
					break
				}
				switch new_all_hall_orders[msg_Id][floor][btn] {
				case NONE:
					new_all_hall_orders[msg_Id][floor][btn] = CONFIRMED
				case NEW:
					new_all_hall_orders[msg_Id][floor][btn] = CONFIRMED
				default:
					break
				}

			case ASSIGNED:
				switch new_all_hall_orders[Id][floor][btn] {
				case NEW:
					new_all_hall_orders[msg_Id][floor][btn] = NEW
				default:
					break
				}
				break

			case COMPLETE:
				switch new_all_hall_orders[Id][floor][btn] {
				case NONE:
					break
				default:
					new_all_hall_orders[msg_Id][floor][btn] = COMPLETE
					new_all_hall_orders[Id][floor][btn] = COMPLETE
				}

			default:
				break
			}
		}
	}
	return new_all_hall_orders
}

func UnassignHallOrders(all_hall_orders [N_ELEVATORS]HallOrders) [N_ELEVATORS]HallOrders {
	new_all_hall_orders := all_hall_orders

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			for index, hall_orders := range new_all_hall_orders {
				switch hall_orders[floor][btn] {
				case ASSIGNED:
					new_all_hall_orders[index][floor][btn] = NEW
					break

				case CONFIRMED:
					new_all_hall_orders[index][floor][btn] = NEW
					break

				default:
					break
				}
			}
		}
	}
	return new_all_hall_orders
}

func OrderDistributionRun(
	Id int,
	cabs_file_name string,
	local_elevator_ch chan elev_struct.Elevator,
	network_Rx_ch chan NetworkMessage,
	peer_update_ch chan peers.PeerUpdate,
	cleared_order_ch chan elev_io.ButtonEvent,
	assigned_orders_local_ch chan<- elev_io.ButtonEvent,
	network_Tx_ch chan<- NetworkMessage,
	clear_local_hall_ch chan<- bool) {

	all_hall_orders := InitializeAllHallOrders()
	available_elevators := SetAvailableElevator(Id, true)
	elevators := InitializeAllElevators()
	stuck_elevators := InitializeAllStuckFlags()

	confirmed_orders := make(chan [N_ELEVATORS]HallOrders, config.BUF_SIZE)
	reset_orders := make(chan [N_ELEVATORS]HallOrders, config.BUF_SIZE)
	hall_lights := make(chan elev_struct.Light, config.BUF_SIZE)

	go ConfirmHallOrders(Id, &all_hall_orders, &available_elevators, confirmed_orders, hall_lights)
	go ResetHallOrders(Id, &all_hall_orders, &available_elevators, reset_orders, hall_lights)

	for {
		select {
		case peer_update := <-peer_update_ch:
			available_elevators = peers.GetAvailable(peer_update, available_elevators)
			if len(peer_update.Lost) > 0 {
				clear_local_hall_ch <- true
				lost_peer_id, _ := strconv.Atoi(peer_update.Lost[0])
				if lost_peer_id != Id {
					all_hall_orders = UnassignHallOrders(all_hall_orders)
					all_hall_orders[lost_peer_id] = InitializeHallOrders()
				} else {
					all_hall_orders = InitializeAllHallOrders()
				}
			}
			if len(peer_update.New) > 0 {
				new_peer_id, _ := strconv.Atoi(peer_update.New)
				if new_peer_id == Id {
					all_hall_orders = InitializeAllHallOrders()
					network_Rx := <-network_Rx_ch
					all_hall_orders[Id] = network_Rx.HallOrders
					clear_local_hall_ch <- true
					all_hall_orders = UnassignHallOrders(all_hall_orders)
				} else {
					all_hall_orders[new_peer_id] = InitializeHallOrders()
					clear_local_hall_ch <- true
					all_hall_orders = UnassignHallOrders(all_hall_orders)
				}
			}

		case local_elevator := <-local_elevator_ch:
			elevators[Id] = local_elevator
			stuck_elevators[Id] = local_elevator.StuckFlag
			cab_calls := elev_struct.GetCabCalls(local_elevator)
			err := backup.SaveCabOrdersToFile(cab_calls, cabs_file_name)
			if err != nil {
				fmt.Println(err)
			}
			if available_elevators[Id] != false {
				all_hall_orders = UpdateHallOrders(Id, local_elevator, all_hall_orders)
			}
			if !stuck_elevators[Id] && local_elevator.StuckFlag {
				all_hall_orders = UnassignHallOrders(all_hall_orders)
				clear_local_hall_ch <- true
			}
			msg := NetworkMessage{local_elevator, all_hall_orders[Id]}
			network_Tx_ch <- msg

		case network_Rx := <-network_Rx_ch:
			if network_Rx.Elevator.ID != Id {
				elevators[network_Rx.Elevator.ID] = network_Rx.Elevator
				if !stuck_elevators[network_Rx.Elevator.ID] && network_Rx.Elevator.StuckFlag {
					all_hall_orders = UnassignHallOrders(all_hall_orders)

					clear_local_hall_ch <- true
				}
				stuck_elevators[network_Rx.Elevator.ID] = network_Rx.Elevator.StuckFlag
				all_hall_orders = UpdateHallOrdersFromNetwork(Id, network_Rx, all_hall_orders)
			}

		case confirmed_orders := <-confirmed_orders:
			all_hall_orders = confirmed_orders
			input := FormatToDistributor(elevators, Id, all_hall_orders, available_elevators, stuck_elevators)
			output := distributor.HallOrdersDistribute(input)
			all_hall_orders = SetOrdersAssigned(Id, all_hall_orders)
			assigned_elevators := UpdateFromDistributor(output, elevators, available_elevators, stuck_elevators)
			assigned_local_orders := assigned_elevators[Id].Requests
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS-1; btn++ {
					if assigned_local_orders[floor][btn] {
						assigned_orders_local_ch <- elev_io.ButtonEvent{floor, elev_io.ButtonType(btn)}
					}
				}
			}

		case cleared_order := <-cleared_order_ch:
			cleared_hall_orders := all_hall_orders
			if cleared_order.Button != elev_io.BT_Cab {
				cleared_hall_orders[Id][cleared_order.Floor][cleared_order.Button] = COMPLETE
			}
			all_hall_orders = cleared_hall_orders
			msg := NetworkMessage{elevators[Id], all_hall_orders[Id]}
			network_Tx_ch <- msg

		case reset_orders := <-reset_orders:
			all_hall_orders = reset_orders

		case hall_lights := <-hall_lights:
			elev_io.SetButtonLamp(hall_lights.Button, hall_lights.Floor, hall_lights.Value)

		default:
		}
	}
}
