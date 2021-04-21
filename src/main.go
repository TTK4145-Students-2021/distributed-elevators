package main

import (
	"flag"
	"fmt"
	"strconv"

	"./controller"
	hw "./hardware"
	"./master"
	"./network/network"
	"./orders"
	. "./types"
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
	sendOnNetworkCh := make(chan NetworkMessage, 200)

	RX :=
		network.RXChannels{
			StateUpdateCh:       make(chan ElevState, 200),
			RegisterOrderCh:     make(chan OrderEvent, 200),
			OrdersFromMasterCh:  make(chan GlobalOrderMap),
			OrderCopyRequestCh:  make(chan bool),
			OrderCopyResponseCh: make(chan GlobalOrderMap, 200),
		}

	masterChannels := master.MasterChannels{
		IsMasterCh:          isMasterCh,
		PeerLostCh:          peerLostCh,
		ToSlavesCh:          sendOnNetworkCh,
		RegisterOrderCh:     RX.RegisterOrderCh,
		StateUpdateCh:       RX.StateUpdateCh,
		OrderCopyResponseCh: RX.OrderCopyResponseCh,
	}

	hwChannels := hw.HardwareChannels{
		FloorSensorCh:       make(chan int),
		StopSensorCh:        make(chan bool),
		ObstructionSensorCh: make(chan bool),
		KeyPressCh:          make(chan ButtonEvent),
	}

	orderChannels := orders.OrderChannels{
		LocalOrderCh:       make(chan OrderMatrix),
		LocalLightCh:       make(chan OrderMatrix),
		ClearedFloorCh:     make(chan int, 200),
		OrdersFromMasterCh: RX.OrdersFromMasterCh,
		OrderCopyRequestCh: RX.OrderCopyRequestCh,
		ToMasterCh:         sendOnNetworkCh,
		KeyPressCh:         hwChannels.KeyPressCh,
	}

	ctrlChannels := controller.ControllerChannels{
		FloorSensorCh:       hwChannels.FloorSensorCh,
		StopSensorCh:        hwChannels.StopSensorCh,
		ObstructionSensorCh: hwChannels.ObstructionSensorCh,
		LocalOrderCh:        orderChannels.LocalOrderCh,
		LocalLightCh:        orderChannels.LocalLightCh,
		ClearedFloorCh:      orderChannels.ClearedFloorCh,
		ToMasterCh:          sendOnNetworkCh,
	}

	fmt.Println("### Starting Elevator ###")
	network.InitNetwork(ID,
		sendOnNetworkCh,
		RX,
		isMasterCh,
		peerLostCh,
	)
	hw.Init(
		"localhost:"+simPort, N_FLOORS,
		hwChannels,
	)

	go master.RunMaster(
		ID,
		masterChannels,
	)
	go controller.StartElevatorController(
		ID,
		ctrlChannels,
	)
	go orders.StartOrderModule(
		ID,
		orderChannels,
	)
	select {}
}
