package distributor

import (
	"elevator-project/config"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string                `json:"behaviour"`
	Floor       int                   `json:"floor"`
	Direction   string                `json:"direction"`
	CabRequests [config.N_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.N_FLOORS][config.N_BUTTONS - 1]bool `json:"hallRequests"`
	States       map[string]HRAElevState                     `json:"states"`
}

func HallOrdersDistribute(input HRAInput) map[string][][config.N_BUTTONS - 1]bool {
	hra_executable := ""

	switch runtime.GOOS {
	case "linux":
		hra_executable = "hall_request_assigner"

	case "windows":
		hra_executable = "hall_request_assigner.exe"

	default:
		panic("OS not supported")
	}

	json_bytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		panic(err)
	}

	ret, err := exec.Command("distributor/"+hra_executable, "-i", string(json_bytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		panic(err)
	}

	output := new(map[string][][config.N_BUTTONS - 1]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		panic(err)
	}
	return *output
}
