package network

import (
	"fmt"
	//"time"

	"../../types"
	"../peers"
)

type RXChannels struct {
	StateUpdateCh       chan types.ElevState      `addr:"stateupdatech"`
	RegisterOrderCh     chan types.OrderEvent     `addr:"registerorderch"`
	OrdersFromMasterCh  chan types.GlobalOrderMap `addr:"ordersfrommasterch"`
	OrderCopyRequestCh  chan bool                 `addr:"ordercopyrequestch"`
	OrderCopyResponseCh chan types.GlobalOrderMap `addr:"ordercopyresponsech"`
}

func InitNetwork(id string, networkSendCh <-chan types.NetworkMessage, rxChannels RXChannels, isMasterCh chan bool, peerLostCh chan string) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	defaultTcpPort := 6942

	portCh := make(chan int, 1)

	//Spawn TCP listen client handler, get assigned port
	go listenAndServe(defaultTcpPort, portCh, rxChannels)

	tcpPort := <-portCh
	fmt.Println("Port ", tcpPort)

	// Start TCP Client handler
	go runTcpClient(id, rxChannels, networkSendCh, peerUpdateCh, isMasterCh, peerLostCh)

	// Start UDP-peer discovery
	go peers.Transmitter(15647, id, tcpPort)
	go peers.Receiver(15647, peerUpdateCh)
}
