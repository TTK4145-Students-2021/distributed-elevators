package master

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	"os/exec"

	// "reflect"
	// "reflect"
	. "../types"
	// "github.com/davecgh/go-spew/spew"
)

/*Types*/
// type SingleElevator struct {
// 	State     State          `json:"state"`
// 	CabOrders [N_FLOORS]bool `json:"cabRequests"`
// }

type SingleElevator struct {
	Behavior  string         `json:"behavior"`
	Floor     int            `json:"floor"`
	Direction string         `json:"direction"`
	CabOrders [N_FLOORS]bool `json:"cabRequests"`
}

type CombinedElevators struct {
	GlobalOrders [N_FLOORS][N_BUTTONS - 1]bool `json:"hallRequests"`
	States       map[string]SingleElevator     `json:"states"`
}

func RunMaster(registerOrder <-chan OrderEvent, updateElevState <-chan State, globalUpdatedOrders chan<- GlobalOrderMap) {
	println("## Running Master ##")
	/* 	channels */
	// reAssign := make(chan bool, 20)
	/* 	variables */
	// e := CombinedElevators{
	// 	GlobalOrders: [N_FLOORS][N_BUTTONS - 1]bool{},
	// 	States:       make(map[string]SingleElevator),
	// }

	gl_orders := [N_FLOORS][N_BUTTONS - 1]bool{}
	gl_states := map[string]SingleElevator{}

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
					completed	bool
					floor		int
					type		[N_BUTTONS]bool


			<- OR global map
				*when master is initiated, it will request the other peers for their copy
				of the global map and OR them together.
				OR'ing will happen here.

			<- redistribute
				*calculate assignment and push to peers
		*/

		case st := <-updateElevState:
			println("M: Got State. ID: ", st.ID)
			_, exist := gl_states[st.ID]

			switch exist {
			case false:
				gl_states[st.ID] = SingleElevator{
					st.Behavior.String(),
					st.Floor,
					st.Direction.String(),
					[N_FLOORS]bool{}}
			case true:
				cab := gl_states[st.ID].CabOrders
				gl_states[st.ID] = SingleElevator{
					st.Behavior.String(),
					st.Floor,
					st.Direction.String(),
					cab}
			}

		case o := <-registerOrder:
			println("M: master got order")
			id := o.ID
			if _, exist := gl_states[id]; !exist {
				println("M: No client with ID: ", o.ID)
				break
			}

			switch o.Order.Button {
			case BT_HallUp, BT_HallDown:
				gl_orders[o.Order.Floor][o.Order.Button] = !o.Completed
			case BT_Cab: //What happenes if order given, but no elevator state present?
				elev := gl_states[id]
				elev.CabOrders[o.Order.Floor] = !o.Completed
				gl_states[id] = elev
			}

			fmt.Println("Reassigning")
			statesAndOrders := CombinedElevators{gl_orders, gl_states}
			updatedOrders := calculateDistribution(statesAndOrders.Json())
			globalUpdatedOrders <- updatedOrders
			// fmt.Println(updatedOrders)
		}
	}
}

func (c CombinedElevators) Json() string {
	// json_byte, _ := json.MarshalIndent(&c, "", "    ")
	json_byte, _ := json.Marshal(&c)
	return string(json_byte)
}

func calculateDistribution(input_json string) GlobalOrderMap {

	// input, err := ioutil.ReadFile("../test.json")
	// check(err)
	out, _ := exec.Command("../hall_request_assigner", "--includeCab", "--input", input_json).Output()

	var assigned_orders GlobalOrderMap
	json.Unmarshal(out, &assigned_orders)

	return assigned_orders
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
