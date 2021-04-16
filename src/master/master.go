package master

import (
	"encoding/json"

	// "io/ioutil"
	"os/exec"

	// "reflect"
	// "reflect"
	. "../types"
	// "github.com/davecgh/go-spew/spew"
)

/*Types*/
type single_elevator struct {
	state       State          `json:"state"`
	cabRequests [N_FLOORS]bool `json:"cabRequests"`
}

type SingleElevator struct {
	Behavior    string         `json:"behavior"`
	Floor       int            `json:"floor"`
	Direction   string         `json:"direction"`
	CabRequests [N_FLOORS]bool `json:"cabRequests"`
}

type CombinedElevators struct {
	GlobalOrders [N_FLOORS][N_BUTTONS - 1]bool `json:"hallRequests"`
	States       map[string]SingleElevator     `json:"states"`
}

func main() {
	/* 	channels */

	/* 	variables */
	// combined := CombinedElevators{
	// 	GlobalOrders: [N_FLOORS][N_BUTTONS - 1]bool{},
	// 	States:       make(map[string]SingleElevator),
	// }

	for {
		select {
		/*
			<- update_single_elevator
				*when a state is sent to be updated
				 struct:
					ID		string
					State	state

			<- order_new
				stuct:
					ID			string
					floor		int
					type		[N_BUTTONS]bool

			<- order_done
				stuct:
					ID			string
					floor		int
					type		[N_BUTTONS]bool


			<- OR global map
				*when master is initiated, it will request the other peers for their copy
				of the global map and OR them together.
				OR'ing will happen here.

			<- redistribute
				*calculate assignment and push to peers
		*/

		}
	}
}

func (c CombinedElevators) Json() string {
	json_byte, _ := json.Marshal(&c)
	return string(json_byte)
}

func calculateDistribution(input_json string) GlobalOrderMap {

	// byte, err := ioutil.ReadFile("../../test.json")
	// check(err)
	out, _ := exec.Command("../../hall_request_assigner", "--includeCab", "--input", input_json).Output()

	var assigned_orders GlobalOrderMap
	json.Unmarshal(out, &assigned_orders)

	return assigned_orders
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
