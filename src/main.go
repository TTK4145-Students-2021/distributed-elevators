package main

import (
	"./controller_fsm"
	// "./hardware_io"
	"./master"
	"./orders"
	. "./types"
	// "./test"
	"fmt"
)

func main() {
	localOrdersCh := make(chan ButtonEvent)
	newOrder := make(chan OrderEvent)
	// doneOrder := make(chan OrderEvent)
	updateElevState := make(chan State)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localOrdersCh, updateElevState)
	go orders.StartOrderModule(localOrdersCh, newOrder)
	go master.RunMaster(newOrder, updateElevState)
	for {
	}
}
