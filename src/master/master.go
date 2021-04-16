package main

import (
	"encoding/json"
	"fmt"
	"strings"

	// "io/ioutil"
	"os/exec"

	// "reflect"
	// "reflect"
	. "../types"
	"github.com/davecgh/go-spew/spew"
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

type AssignReqInput struct {
	HallRequests [N_FLOORS][N_BUTTONS - 1]bool `json:"hallRequests"`
	States       map[string]SingleElevator     `json:"states"`
}
type OrderMatrix [N_FLOORS][N_BUTTONS]bool

type GlobalOrderMap map[string]OrderMatrix

func main() {
	/* 	channels */

	/* 	variables */

	/* 	map["id"] SingleElevator
	   	global_orders [N_FLOOR][2] `json:"hallRequests"`

	*/
	global_orders := [N_FLOORS][N_BUTTONS - 1]bool{}
	e_states := make(map[string]SingleElevator)

	spew.Dump("spew\n")
	fmt.Println("fmt")
	elev := SingleElevator{"moving", 2, "up", [N_FLOORS]bool{false, false, true, true}}

	// spew.Dump(elev)
	e_states["heis1"] = elev

	// spew.Dump(e_states)

	global_orders[1][0] = true
	global_orders[3][1] = true

	input := AssignReqInput{global_orders, e_states}
	global_order_map := calculateDistribution(input.Json())

	fmt.Println(global_order_map)

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

func (in AssignReqInput) Json() string {
	json_byte, _ := json.Marshal(&in)
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

func (mat OrderMatrix) String() string {
	var s []string
	for i := range mat {
		s = append(s, fmt.Sprintf("f%d: ", i+1))
		for _, n := range mat[i] {
			b2i := map[bool]int8{false: 0, true: 1}
			s = append(s, fmt.Sprintf("%d   ", b2i[n]))
		}
		s = append(s, "\n")
	}
	return strings.Join(s, "")
}

func (m GlobalOrderMap) String() string {
	var s []string
	for name, mat := range m {
		// fmt.Printf("teller")
		s = append(s, fmt.Sprintln(name+":"))
		s = append(s, fmt.Sprint(mat))
	}
	return strings.Join(s, "")
}
