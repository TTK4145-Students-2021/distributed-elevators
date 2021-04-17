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
	localUpdatedOrders := make(chan OrderMatrix)
	localUpdatedLights := make(chan OrderMatrix)
	registerOrder := make(chan OrderEvent, 10)
	completedOrder := make(chan int, 200)
	// doneOrder := make(chan OrderEvent)
	updateElevState := make(chan State)
	globalUpdatedOrders := make(chan GlobalOrderMap)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localUpdatedOrders, localUpdatedLights, updateElevState, completedOrder)
	go orders.StartOrderModule(localUpdatedOrders, localUpdatedLights, registerOrder, globalUpdatedOrders, completedOrder)
	go master.RunMaster(registerOrder, updateElevState, globalUpdatedOrders)
	for {
	}
}
