package main

import (
	"./controller_fsm"
	// "./hardware_io"
	"./master"
	"./orders"
	. "./types"
	// "./test"
	"fmt"
	"flag"
	"./network/network"
)

func main() {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	globalUpdatedOrdersChannel := make(chan GlobalOrderMap)
	updateElevStateChannel := make(chan State, 200)
	RXChannels := RXChannels{StateCh: updateElevStateChannel,
	GlobalOrdersCh: globalUpdatedOrdersChannel}
	networkSendCh := make(chan NetworkMessage, 200)
	network.NetworkTest(id, networkSendCh, RXChannels)
	

	localUpdatedOrders := make(chan OrderMatrix)
	localUpdatedLights := make(chan OrderMatrix)
	registerOrder := make(chan OrderEvent, 200)
	completedOrder := make(chan int, 200)
	// doneOrder := make(chan OrderEvent)
	

	orderMergeCh := make(chan GlobalOrderMap)
	iAmMasterCh := make(chan bool)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localUpdatedOrders, localUpdatedLights, networkSendCh, completedOrder)
	go orders.StartOrderModule(localUpdatedOrders, localUpdatedLights, registerOrder, globalUpdatedOrdersChannel, completedOrder, orderMergeCh)
	go master.ListenForMasterUpdate(iAmMasterCh, registerOrder, updateElevStateChannel, networkSendCh, orderMergeCh) //make a struct for channels
	iAmMasterCh <- true

	for {
		select{}
		/*select{
	case a:= <-updateElevStateChannel:
		fmt.Println("Got state update: ",a)
	}*/
}
}
