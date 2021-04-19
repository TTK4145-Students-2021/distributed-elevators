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
	"./network/network"
	"flag"
)

func main() {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	iAmMasterCh := make(chan bool)

	stateUpdateCh := make(chan State, 200)
	registerOrderCh := make(chan OrderEvent, 200)
	orderCopyRequestCh := make(chan bool)

	ordersFromMasterCh := make(chan GlobalOrderMap)
	orderCopyResponseCh := make(chan GlobalOrderMap)

	RXChannels :=
		RXChannels{
			StateUpdateCh:       stateUpdateCh,
			RegisterOrderCh:     registerOrderCh,
			OrdersFromMasterCh:  ordersFromMasterCh,
			OrderCopyRequestCh:  orderCopyRequestCh,
			OrderCopyResponseCh: orderCopyResponseCh,
		}

	networkSendCh := make(chan NetworkMessage, 200)
	network.InitNetwork(id, networkSendCh, RXChannels, iAmMasterCh)

	//internal
	localOrderCh := make(chan OrderMatrix)
	localLightCh := make(chan OrderMatrix)
	clearedFloorCh := make(chan int, 200)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(
		localOrderCh,
		localLightCh,
		clearedFloorCh,
		networkSendCh)
	go orders.StartOrderModule(
		localOrderCh,
		localLightCh,
		clearedFloorCh,
		networkSendCh,
		ordersFromMasterCh,
		orderCopyRequestCh)
	go master.RunMaster(
		iAmMasterCh,
		registerOrderCh,
		stateUpdateCh,
		networkSendCh,
		orderCopyResponseCh) //make a struct for channels
	iAmMasterCh <- true

	for {
		select {}
		/*select{
		case a:= <-updateElevStateChannel:
			fmt.Println("Got state update: ",a)
		}*/
	}
}
