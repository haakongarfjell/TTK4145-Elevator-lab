package backup

import (
	"elevator-project/config"
	"elevator-project/local_elevator/elev_io"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const (
	N_FLOORS  int = config.N_FLOORS
	N_BUTTONS     = config.N_BUTTONS
	BT_CAB    int = elev_io.BT_Cab
)

func ReadCabCallsFromFile(file_name string) ([N_FLOORS]bool, error) {
	data, err := ioutil.ReadFile(file_name)

	if err != nil {
		var none [N_FLOORS]bool
		return none, err
	}

	var cab_calls [N_FLOORS]bool
	err = json.Unmarshal(data, &cab_calls)

	if err != nil {
		var none [N_FLOORS]bool
		return none, err
	}
	return cab_calls, nil
}

func InitCabCalls(cab_button_ch chan<- elev_io.ButtonEvent, file_name string) {
	cab_calls, err := ReadCabCallsFromFile(file_name)

	if err != nil {
		fmt.Println(err)
	}

	for floor := 0; floor < N_FLOORS; floor++ {
		if cab_calls[floor] {
			cab_button_ch <- elev_io.ButtonEvent{floor, elev_io.ButtonType(BT_CAB)}
		}
	}
}

func SaveCabOrdersToFile(cab_calls [N_FLOORS]bool, file_name string) error {
	old_cab_calls, err := ReadCabCallsFromFile(file_name)

	if err != nil {
		return err
	}

	if cab_calls != old_cab_calls {
		data, err := json.Marshal(cab_calls)

		if err != nil {
			return err
		}

		err = ioutil.WriteFile(file_name, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
