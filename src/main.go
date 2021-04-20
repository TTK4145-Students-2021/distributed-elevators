package main

import (
	"strconv"
	"time"

	"./controller"
	hw "./hardware"

	"./master"
	"./network/network"
	"./orders"
	t "./types"
	"flag"
	"fmt"
)

func main() {
	var ID string
	var simPort string
	flag.StringVar(&ID, "id", "", "id of this peer")
	flag.StringVar(&simPort, "simport", "15657", "port for simulator")
	flag.Parse()

	_, err := strconv.Atoi(ID)
	if err != nil {
		println("ERROR: ID missing or non-integer")
		return
	}
	var simAddr string = "localhost:" + simPort

	// iAmMasterCh := make(chan bool)
	isMasterCh := make(chan bool) //Seperate master channels for testing

	peerLostCh := make(chan string, 200)

	stateUpdateCh := make(chan t.ElevState, 200)
	registerOrderCh := make(chan t.OrderEvent, 200)
	orderCopyRequestCh := make(chan bool)

	ordersFromMasterCh := make(chan t.GlobalOrderMap)
	orderCopyResponseCh := make(chan t.GlobalOrderMap)

	RXChannels :=
		network.RXChannels{
			StateUpdateCh:       stateUpdateCh,
			RegisterOrderCh:     registerOrderCh,
			OrdersFromMasterCh:  ordersFromMasterCh,
			OrderCopyRequestCh:  orderCopyRequestCh,
			OrderCopyResponseCh: orderCopyResponseCh,
		}

	networkSendCh := make(chan t.NetworkMessage, 200)
	network.InitNetwork(ID, networkSendCh, RXChannels, isMasterCh, peerLostCh)
	//internal
	localOrderCh := make(chan t.OrderMatrix)
	localLightCh := make(chan t.OrderMatrix)
	clearedFloorCh := make(chan int, 200)

	fmt.Println("### Starting Elevator ###")
	time.Sleep(1 * time.Second)
	hw.Init(simAddr, t.N_FLOORS)
	go master.RunMaster(
		ID,
		isMasterCh,
		registerOrderCh,
		stateUpdateCh,
		networkSendCh,
		orderCopyResponseCh,
		peerLostCh) //make a struct for channels
	go controller.StartElevatorController(
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
