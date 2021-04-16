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

func RunMaster(newOrder <-chan OrderEvent, updateElevState <-chan State) {
	println("## Running Master ##")
	/* 	channels */

	/* 	variables */
	e := CombinedElevators{
		GlobalOrders: [N_FLOORS][N_BUTTONS - 1]bool{},
		States:       make(map[string]SingleElevator),
	}

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

		case state := <-updateElevState: //new_state
			/* Shitty kode, bør skrives om for lesbarhet */
			println("M: Got State: ID: ", state.ID)
			_, exists := e.States[state.ID]
			if !exists {
				e.States[state.ID] = SingleElevator{
					state.Behavior.String(),
					state.Floor,
					state.Direction.String(),
					[N_FLOORS]bool{},
				}
			} else {
				e.States[state.ID] = SingleElevator{
					state.Behavior.String(),
					state.Floor,
					state.Direction.String(),
					e.States[state.ID].CabRequests,
				}
			}

		case a := <-newOrder:
			println("M: master got order")
			client_elev, ok := e.States[a.ID]
			if !ok {
				println("M: No client with ID: ", a.ID)
				return
			}

			switch a.Order.Button {
			case BT_Cab:
				arr := e.States[a.ID].CabRequests
				arr[a.Order.Floor] = true
				e.States[a.ID] = 
			}
			println("M: ", e.States[a.ID].CabRequests[a.Order.Floor])

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
