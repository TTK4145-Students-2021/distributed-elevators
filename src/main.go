package main

import (
	"./controller"
	hw "./hardware"
	"strconv"

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
	flag.StringVar(&ID, "id", "", "id of this elevator")
	flag.StringVar(&simPort, "simport", "15657", "port for simulator")
	flag.Parse()

	_, err := strconv.Atoi(ID)
	if err != nil {
		println("ERROR: ID missing or non-integer")
		return
	}

	isMasterCh := make(chan bool)
	peerLostCh := make(chan string, 200)

	RXChannels :=
		network.RXChannels{
			StateUpdateCh:       make(chan t.ElevState, 200),
			RegisterOrderCh:     make(chan t.OrderEvent, 200),
			OrdersFromMasterCh:  make(chan t.GlobalOrderMap),
			OrderCopyRequestCh:  make(chan bool),
			OrderCopyResponseCh: make(chan t.GlobalOrderMap),
		}

	networkSendCh := make(chan t.NetworkMessage, 200)
	network.InitNetwork(ID, networkSendCh, RXChannels, isMasterCh, peerLostCh)

	masterChannels := master.MasterChannels{
		IsMasterCh: chan bool,
		PeerLostCh: chan string,
		ToSlavesCh: chan NetworkMessage,
		RegisterOrderCh: chan OrderEvent,
		StateUpdateCh: <-chan ElevState,
		OrderCopyResponseCh: chan GlobalOrderMap,
	}


	hwChannels := hw.hardwareChannels{
		FloorSensorCh:       make(chan int),
		StopSensorCh:        make(chan bool),
		ObstructionSensorCh: make(chan bool),
		KeyPressCh:          make(chan t.ButtonEvent),
	}

	orderChannels := orders.OrderChannels{
		LocalOrderCh:       make(chan t.OrderMatrix),
		LocalLightCh:       make(chan t.OrderMatrix),
		ClearedFloorCh:     make(chan int, 200),
		OrdersFromMasterCh: make(chan t.GlobalOrderMap),
		OrderCopyRequestCh: make(chan bool),
		ToMasterCh:         networkSendCh,
		KeyPressCh:         hwChannels.KeyPressCh,
	}

	ctrChannels := controller.ControllerChannels{
		FloorSensorCh:       hwChannels.FloorSensorCh,
		StopSensorCh:        hwChannels.StopSensorCh,
		ObstructionSensorCh: hwChannels.ObstructionSensorCh,
		LocalOrderCh:        orderChannels.LocalOrderCh,
		LocalLightCh:        orderChannels.LocalLightCh,
		ClearedFloorCh:      orderChannels.ClearedFloorCh,
		ToMasterCh:          networkSendCh,
	}


	fmt.Println("### Starting Elevator ###")
	hw.Init(
		"localhost:"+simPort, t.N_FLOORS,
		hwChannels,
	)

	go master.RunMaster(
		ID,
		isMasterCh,
		registerOrderCh,
		stateUpdateCh,
		networkSendCh,
		orderCopyResponseCh,
		peerLostCh,
	)
	go controller.StartElevatorController(
		ID,
		ctrChannels,
	)
	go orders.StartOrderModule(
		ID,
		orderChannels,
	)
	select {}
}
