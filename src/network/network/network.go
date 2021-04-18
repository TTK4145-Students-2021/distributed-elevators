package network

import (
	"fmt"
	//"time"

	"../TCPmsg"
	"../peers"
	"../../types"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func InitNetwork(id string, networkSendCh <-chan types.NetworkMessage, rxChannels types.RXChannels, isMasterUpdate chan bool) {

	

	peerUpdateCh := make(chan peers.PeerUpdate)
	tcpPort := 8080
	//tcpMsgCh := make(chan types.NetworkMessage, 200)

	tcpPort = runTCPServerAndClient(rxChannels, networkSendCh, peerUpdateCh, tcpPort)
	go runUDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)

}

func runTCPServerAndClient(rxCh types.RXChannels, tcpMsgCh <-chan types.NetworkMessage, peerUpdateCh <-chan peers.PeerUpdate, tcpPort int) int {

	portCh := make(chan int, 1)
	server := TCPmsg.NewServer(rxCh)

	//Spawn TCP listen client handler, get assigned port
	go server.ListenAndServe(tcpPort, portCh)
	tcpPort = <-portCh
	fmt.Println("Port ", tcpPort)

	// Start TCP Client handler
	go TCPmsg.ClientHandler(tcpMsgCh, peerUpdateCh)
	return tcpPort
}

func runUDPServer(id string, tcpPort int, isMasterUpdate chan bool, peerUpdateCh chan<- peers.PeerUpdate) {
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, tcpPort, isMasterUpdate, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
}
