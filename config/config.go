package config

import "time"

const (
	N_FLOORS  int = 4
	N_BUTTONS int = 3
	N_ELEVATORS int = 3
	BETWEEN_FLOOR int = -1

	T_DOOR  time.Duration = 3
	T_STUCK time.Duration = 5
	T_SLEEP time.Duration = 30

	P_PEERS int = 15601
	P_BCAST int = 16500

	BUF_SIZE int = 10
)