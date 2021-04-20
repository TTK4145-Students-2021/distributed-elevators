package network

import (
	"fmt"
	//"time"

	"../../types"
	"../peers"
	kcp "github.com/xtaci/kcp-go"
	"net"
)

type RXChannels struct {
	StateUpdateCh       chan types.ElevState      `addr:"stateupdatech"`
	RegisterOrderCh     chan types.OrderEvent     `addr:"registerorderch"`
	OrdersFromMasterCh  chan types.GlobalOrderMap `addr:"ordersfrommasterch"`
	OrderCopyRequestCh  chan bool                 `addr:"ordercopyrequestch"`
	OrderCopyResponseCh chan types.GlobalOrderMap `addr:"ordercopyresponsech"`
}

const peerDetectionPort = 15647

var serverPort = 6942

func InitNetwork(id string, networkSendCh <-chan types.NetworkMessage, rxChannels RXChannels, isMasterCh chan bool, peerLostCh chan string) {

	peerUpdateCh := make(chan peers.PeerUpdate)

	serverPort, connection := getAvaileblePort(serverPort)
	fmt.Println("Server listening on port:", serverPort)

	// Start peer discovery
	go peers.Transmitter(peerDetectionPort, id, serverPort)
	go peers.Receiver(peerDetectionPort, peerUpdateCh)

	go runServer(serverPort, connection, rxChannels)
	go runClient(id, rxChannels, networkSendCh, peerUpdateCh, isMasterCh, peerLostCh)

}

func getAvaileblePort(port int) (int, net.Listener) {
	var connection net.Listener
	for {
		var err error
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		connection, err = kcp.Listen(addr)
		if err != nil {
			fmt.Println("Listen err ", err)
			port++
		} else {
			break
		}
	}
	return port, connection
}
