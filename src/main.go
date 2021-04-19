package main

import (
	"./controller_fsm"
	"fmt"
	// "./hardware_io"
	"./master"
	// "./network"
	"./orders"
	. "./types"
	// "./test"
)

func main() {
	localUpdatedOrders := make(chan OrderMatrix)
	localUpdatedLights := make(chan OrderMatrix)
	registerOrder := make(chan OrderEvent, 200)
	completedOrder := make(chan int, 200)
	// doneOrder := make(chan OrderEvent)
	updateElevState := make(chan State, 200)
	globalUpdatedOrders := make(chan GlobalOrderMap)

	requestClientOrderCopy := make(chan bool, 200)
	orderMergeCh := make(chan GlobalOrderMap, 200)
	iAmMasterCh := make(chan bool)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localUpdatedOrders, localUpdatedLights, updateElevState, completedOrder)
	go orders.StartOrderModule(localUpdatedOrders, localUpdatedLights, registerOrder, globalUpdatedOrders, completedOrder, orderMergeCh, requestClientOrderCopy)
	go master.RunMaster(iAmMasterCh, registerOrder, updateElevState, globalUpdatedOrders, orderMergeCh, requestClientOrderCopy) //make a struct for channels
	iAmMasterCh <- true

	// go network.StartNetworkShit()

	for {
	}
}
