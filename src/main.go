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
	registerOrder := make(chan OrderEvent, 200)
	completedOrder := make(chan int, 200)
	// doneOrder := make(chan OrderEvent)
	updateElevState := make(chan State, 200)
	globalUpdatedOrders := make(chan GlobalOrderMap)

	orderMergeCh := make(chan GlobalOrderMap)
	iAmMasterCh := make(chan bool)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localUpdatedOrders, localUpdatedLights, updateElevState, completedOrder)
	go orders.StartOrderModule(localUpdatedOrders, localUpdatedLights, registerOrder, globalUpdatedOrders, completedOrder, orderMergeCh)
	go master.ListenForMasterUpdate(iAmMasterCh, registerOrder, updateElevState, globalUpdatedOrders, orderMergeCh) //make a struct for channels
	iAmMasterCh <- true

	for {
	}
}
