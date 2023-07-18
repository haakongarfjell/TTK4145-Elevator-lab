package main

import (
	"elevator-project/local_elevator/backup"
	"elevator-project/config"
	"elevator-project/local_elevator/elev_io"
	"elevator-project/local_elevator/elev_struct"
	"elevator-project/local_elevator/single_elevator"
	"elevator-project/network/bcast"
	"elevator-project/network/peers"
	"elevator-project/orders"
	"flag"
	"strconv"
)

func main() {

	var id_string string
	flag.StringVar(&id_string, "id", "", "id of this peer")
	var port string
	flag.StringVar(&port, "port", "", "port of this elevator")
	flag.Parse()

	// Differentiating between cabFiles is only necesseary when working on simulator
	cabs_file_name := "cabCalls.json"

	id_int, _ := strconv.Atoi(id_string)

	elev_io.Init("localhost:"+port, config.N_FLOORS)

	for elev_io.GetFloor() == config.BETWEEN_FLOOR {
		elev_io.SetMotorDirection(elev_io.MD_Down)
	}
	elev_io.SetMotorDirection(elev_io.MD_Stop)

	// Driver
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_buttons := make(chan elev_io.ButtonEvent)

	// Local elevator
	local_elevator := make(chan elev_struct.Elevator)
	cleared_order := make(chan elev_io.ButtonEvent, config.BUF_SIZE)
	assigned_orders_local := make(chan elev_io.ButtonEvent, config.BUF_SIZE)
	clear_local_hall := make(chan bool, config.BUF_SIZE)

	// Network
	network_Tx := make(chan orders.NetworkMessage)
	network_Rx := make(chan orders.NetworkMessage)
	peer_update := make(chan peers.PeerUpdate)
	peer_Tx_enable := make(chan bool)

	go elev_io.PollFloorSensor(drv_floors)
	go elev_io.PollObstructionSwitch(drv_obstr)
	go elev_io.PollButtons(drv_buttons)

	go peers.Transmitter(config.P_PEERS, id_string, peer_Tx_enable)
	go peers.Receiver(config.P_PEERS, peer_update)
	go bcast.Transmitter(config.P_BCAST, network_Tx)
	go bcast.Receiver(config.P_BCAST, network_Rx)

	backup.InitCabCalls(assigned_orders_local, cabs_file_name)

	go single_elevator.SingleElevatorRun(id_int, drv_floors, drv_obstr, drv_buttons, clear_local_hall, cleared_order, local_elevator, assigned_orders_local)
	go orders.OrderDistributionRun(id_int, cabs_file_name, local_elevator, network_Rx, peer_update, cleared_order, assigned_orders_local, network_Tx, clear_local_hall)

	for {

	}
}