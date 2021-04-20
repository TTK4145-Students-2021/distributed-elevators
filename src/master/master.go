package master

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	"os/exec"

	. "../types"
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

// func ListenForMasterUpdate(iAmMasterCh <-chan bool, registerOrder <-chan OrderEvent, updateElevState <-chan State, globalUpdatedOrders chan<- GlobalOrderMap, orderMergeCh <-chan GlobalOrderMap, requestClientOrderCopy chan<- bool) {
// 	for {
// 		select {
// 		case <-iAmMasterCh:
// 			go RunMaster(iAmMasterCh, registerOrder, updateElevState, globalUpdatedOrders, orderMergeCh, requestClientOrderCopy)
// 			return
// 		}
// 	}
// }

func RunMaster(
	ID string,
	iAmMasterCh <-chan bool,
	registerOrderCh <-chan OrderEvent,
	stateUpdateCh <-chan State,
	toSlavesCh chan<- NetworkMessage,
	orderCopyResponseCh <-chan GlobalOrderMap,
	peerLostCh <-chan string,
) {
	println("## Running Master ##")

	hallOrders := [N_FLOORS][N_BUTTONS - 1]bool{}
	allElevatorStates := map[string]SingleElevator{}

	/* REQUEST ALL ORDER LIST FROM PEERS HERE*/
	requestOrderCopy := NetworkMessage{
		Data:       true,
		Receipient: All,
		ChAddr:     "ordercopyrequestch",
	}

	toSlavesCh <- requestOrderCopy

	for {
		select {
		case st := <-stateUpdateCh:
			shouldReAssign := false
			elevator, exist := allElevatorStates[st.ID]

			CabOrders := [N_FLOORS]bool{}
			if exist {
				CabOrders = elevator.CabOrders
				shouldReAssign = elevator.available != st.Available
			}
			allElevatorStates[st.ID] = SingleElevator{
				st.Behavior.String(),
				st.Floor,
				st.Direction.String(),
				st.Available,
				CabOrders}

			if shouldReAssign {
				updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
				toSlavesCh <- updatedOrders
			}

		case order := <-registerOrderCh:
			//debug
			// a := map[bool]string{false: "new", true: "completed"}
			// println("M: master got", a[order.Completed])
			// //debug^

			id := order.ElevID
			if _, exist := allElevatorStates[id]; !exist { //What happenes if order given, but no elevator state present?
				println("M: No client with ID: ", order.ElevID)
				break
			}
			for _, o := range order.Orders {
				switch o.Button {
				case BT_HallUp, BT_HallDown:
					hallOrders[o.Floor][o.Button] = !order.Completed
				case BT_Cab: //What happenes if order given, but no elevator state present?
					elev := allElevatorStates[id]
					elev.CabOrders[o.Floor] = !order.Completed
					allElevatorStates[id] = elev
				}
			}
			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			toSlavesCh <- updatedOrders

		case iAmMaster := <-iAmMasterCh:
			if iAmMaster {
				toSlavesCh <- requestOrderCopy
				println("requesting order copy")
			} else {
				fmt.Println("MASTER SHUTTING DOWN ON ELEVATOR " + ID)
			sleep:
				for {
					select {
					case iAmMaster := <-iAmMasterCh:
						if iAmMaster {
							toSlavesCh <- requestOrderCopy
							fmt.Println("MASTER WAKING UP")
							break sleep
						}
					}
				}
			}
		case lostPeer := <-peerLostCh:
			elevator, exist := allElevatorStates[lostPeer]

			if !exist {
				break
			} else {
				elevator.available = false
				allElevatorStates[lostPeer] = elevator
				updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
				toSlavesCh <- updatedOrders
			}

		case orderCopy := <-orderCopyResponseCh: //rename to mergeResponse?
			/*
				<- OR global map
					*when master is initiated, it will request the other peers for their copy
					of the global map and OR them together.
					OR'ing will happen here.
			*/
			fmt.Println("M: got order copy response ")
			for elevID, orderMatrix := range orderCopy {
				for f, row := range orderMatrix {
					for b, isOrder := range row {
						switch ButtonType(b) {
						case BT_HallUp, BT_HallDown:
							hallOrders[f][b] = hallOrders[f][b] || isOrder
						case BT_Cab:
							elevator, exist := allElevatorStates[elevID]
							if !exist {
								cabOrders := [N_FLOORS]bool{}
								cabOrders[f] = isOrder
								allElevatorStates[elevID] = SingleElevator{
									"idle",
									0,
									"down",
									true,
									cabOrders}
							} else {
								elevator.CabOrders[f] = elevator.CabOrders[f] || isOrder
								allElevatorStates[elevID] = elevator
							}
						}
					}
				}
			}
			updatedOrders := reAssignOrders(hallOrders, allElevatorStates)
			toSlavesCh <- updatedOrders
		}
	}
}

func reAssignOrders(hallOrders [N_FLOORS][N_BUTTONS - 1]bool, allElevatorStates map[string]SingleElevator) NetworkMessage {
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
	// 	// HANDLES WHEN INPUTMAP EMPTY -> makes it so orders are assigned
	// 	fmt.Println("M: Shiii, we got an empty inputmap in reAssignOrders") //remove
	// 	inputmap = allElevatorStates
	// 	unavailable = nil
	// }

	//calculationg distribution based on input
	jsonInput := CombinedElevators{hallOrders, inputmap}.Json()
	orderList := calculateDistribution(jsonInput)

	//adding non-assigned elevators back to the list
	for _, elevID := range unavailable {
		orders := OrderMatrix{}
		for floor := range orders {
			orders[floor][BT_Cab] = allElevatorStates[elevID].CabOrders[floor]
			fmt.Println("unavailable is ", unavailable)
			orderList[elevID] = orders
		}
	}
	updatedOrders := NetworkMessage{
		Data:       orderList,
		Receipient: All,
		ChAddr:     "ordersfrommasterch"}

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
