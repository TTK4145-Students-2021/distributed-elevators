package network

import (
	"fmt"
	//"time"

	"../../types"
	"../TCPmsg"
	"../peers"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func InitNetwork(id string, networkSendCh <-chan types.NetworkMessage, rxChannels types.RXChannels, isMasterUpdate chan bool, peerLostCh chan string) {

	peerUpdateCh := make(chan peers.PeerUpdate)
	defaultTcpPort := 6942

	tcpPort := runTCPServerAndClient(id, rxChannels, networkSendCh, peerUpdateCh, defaultTcpPort, isMasterUpdate, peerLostCh)
	go runUDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)
}

func runTCPServerAndClient(id string, rxCh types.RXChannels, tcpMsgCh <-chan types.NetworkMessage, peerUpdateCh chan peers.PeerUpdate, tcpPort int, isMaster chan<- bool, peerLostCh chan<- string) int {

	portCh := make(chan int, 1)

	//Spawn TCP listen client handler, get assigned port
	go TCPmsg.ListenAndServe(tcpPort, portCh, rxCh)
	tcpPort = <-portCh
	fmt.Println("Port ", tcpPort)

	// Start TCP Client handler
	go TCPmsg.ClientHandler(id, rxCh, tcpMsgCh, peerUpdateCh, isMaster, peerLostCh)
	return tcpPort
}

func runUDPServer(id string, tcpPort int, isMasterUpdate <-chan bool, peerUpdateCh chan<- peers.PeerUpdate) {
	go peers.Transmitter(15647, id, tcpPort)
	go peers.Receiver(15647, peerUpdateCh)
}
