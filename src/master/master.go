package master

import (
	"encoding/json"
	"fmt"
	// "io/ioutil"
	. "../types"
	"os/exec"
	// "github.com/davecgh/go-spew/spew"
)

/*Types*/
type CombinedElevators struct {
	GlobalOrders [N_FLOORS][N_BUTTONS - 1]bool `json:"hallRequests"`
	States       map[string]SingleElevator     `json:"states"`
}

type SingleElevator struct {
	Behavior  string `json:"behavior"`
	Floor     int    `json:"floor"`
	Direction string `json:"direction"`
	available bool
	CabOrders [N_FLOORS]bool `json:"cabRequests"`
}

/* channels */

func ListenForMasterUpdate(iAmMasterCh <-chan bool, registerOrder <-chan OrderEvent, updateElevState <-chan State, globalUpdatedOrders chan<- GlobalOrderMap, orderMergeCh <-chan GlobalOrderMap) {
	for {
		select {
		case <-iAmMasterCh:
			go RunMaster(iAmMasterCh, registerOrder, updateElevState, globalUpdatedOrders, orderMergeCh)
			return
		}
	}
}

func RunMaster(iAmMasterCh <-chan bool, registerOrder <-chan OrderEvent, updateElevState <-chan State, globalUpdatedOrders chan<- GlobalOrderMap, orderMergeCh <-chan GlobalOrderMap) {
	println("## Running Master ##")

	hallOrders := [N_FLOORS][N_BUTTONS - 1]bool{}
	allElevatorStates := map[string]SingleElevator{}

	/* REQUEST ALL ORDER LIST FROM PEERS HERE*/

	for {
		select {
		/*

			<- OR global map
				*when master is initiated, it will request the other peers for their copy
				of the global map and OR them together.
				OR'ing will happen here.
		*/

		case st := <-updateElevState:
			println("M: Got State. ID: ", st.ID)
			elevator, exist := allElevatorStates[st.ID]

			CabOrders := [N_FLOORS]bool{}
			if exist {
				CabOrders = elevator.CabOrders
			}
			allElevatorStates[st.ID] = SingleElevator{
				st.Behavior.String(),
				st.Floor,
				st.Direction.String(),
				st.Available,
				CabOrders}

		case order := <-registerOrder:
			//debug
			a := map[bool]string{false: "new", true: "completed"}
			println("M: master got", a[order.Completed])
			//debug^
			id := order.ElevID
			if _, exist := allElevatorStates[id]; !exist { //What happenes if order given, but no elevator state present?
				println("M: No client with ID: ", order.ElevID)
				break
			}
			switch order.Order.Button {
			case BT_HallUp, BT_HallDown:
				hallOrders[order.Order.Floor][order.Order.Button] = !order.Completed
			case BT_Cab: //What happenes if order given, but no elevator state present?
				elev := allElevatorStates[id]
				elev.CabOrders[order.Order.Floor] = !order.Completed
				allElevatorStates[id] = elev
			}

			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			globalUpdatedOrders <- updatedOrders

		case iAmMaster := <-iAmMasterCh:
			if iAmMaster {
				//REQUEST ORDER LIST FROM PEERS
			} else {
				go ListenForMasterUpdate(iAmMasterCh, registerOrder, updateElevState, globalUpdatedOrders, orderMergeCh)
				return
			}

		}
	}
}

func reAssignOrders(hallOrders [N_FLOORS][N_BUTTONS - 1]bool, allElevatorStates map[string]SingleElevator) GlobalOrderMap {
	//removing non available elevators from input
	var unavailable []string
	inputmap := make(map[string]SingleElevator)
	for elevID, elevState := range allElevatorStates {
		if elevState.available == false {
			unavailable = append(unavailable, elevID)
		} else {
			inputmap[elevID] = elevState
		}
	}
	// if len(inputmap) == 0 {
	// 	//TODO: HANDLE WHEN INPUTMAP EMPTY
	// 	fmt.Println("M: Shiii, we got an empty inputmap in reAssignOrders") //remove
	// 	inputmap = allElevatorStates
	// 	unavailable = nil
	// }

	//calculationg distribution based on input
	jsonInput := CombinedElevators{hallOrders, inputmap}.Json()
	updatedOrders := calculateDistribution(jsonInput)

	//adding non-assigned elevators back to the list
	for _, elevID := range unavailable {
		orders := OrderMatrix{}
		for floor := range orders {
			orders[floor][BT_Cab] = allElevatorStates[elevID].CabOrders[floor]
			fmt.Println("unavailable is ", unavailable)
			updatedOrders[elevID] = orders
		}
	}
	return updatedOrders
}

func (c CombinedElevators) Json() string {
	// json_byte, _ := json.MarshalIndent(&c, "", "    ")
	json_byte, _ := json.Marshal(&c)
	return string(json_byte)
}

func calculateDistribution(input_json string) GlobalOrderMap {

	// input, err := ioutil.ReadFile("../test.json")
	out, _ := exec.Command("../hall_request_assigner", "--includeCab", "--input", input_json).Output()

	assigned_orders := make(GlobalOrderMap)
	json.Unmarshal(out, &assigned_orders)

	return assigned_orders
}
