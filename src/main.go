package main

import (
	"strconv"
	"time"

	"./controller_fsm"
	hw "./hardware_io"

	// "./hardware_io"
	"./master"
	// "./network"
	"./orders"
	. "./types"

	// "./test"
	"fmt"

	"flag"

	"./network/network"
)

func main() {
	var ID string
	var simPort string
	flag.StringVar(&ID, "id", "", "id of this peer")
	flag.StringVar(&simPort, "simPort", "15657", "id of this peer")
	flag.Parse()

	_, err := strconv.Atoi(ID)
	if err != nil {
		println("ERROR: ID missing or not integer")
		return
	}
	var simAddr string = "localhost:" + simPort
	hw.Init(simAddr, N_FLOORS)

	// iAmMasterCh := make(chan bool)
	isMasterCh := make(chan bool) //Seperate master channels for testing

	peerLostCh := make(chan string, 200)

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
	network.InitNetwork(ID, networkSendCh, RXChannels, isMasterCh, peerLostCh)
	//internal
	localOrderCh := make(chan OrderMatrix)
	localLightCh := make(chan OrderMatrix)
	clearedFloorCh := make(chan int, 200)

	fmt.Println("### Starting Elevator ###")
	go master.RunMaster(
		ID,
		isMasterCh,
		registerOrderCh,
		stateUpdateCh,
		networkSendCh,
		orderCopyResponseCh) //make a struct for channels
	time.Sleep(1 * time.Second)

	go controller_fsm.StartElevatorController(
		ID,
		localOrderCh,
		localLightCh,
		clearedFloorCh,
		networkSendCh,
	)
	go orders.StartOrderModule(
		ID,
		localOrderCh,
		localLightCh,
		clearedFloorCh,
		networkSendCh,
		ordersFromMasterCh,
		orderCopyRequestCh)

	for {
		select {}
	}
}
