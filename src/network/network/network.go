package network

import (
	"fmt"
	"time"

	"../TCPmsg"
	"../peers"
	"../../types"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func NetworkTest(id string) {

	

	peerUpdateCh := make(chan peers.PeerUpdate)
	stateMsgCh := make(chan types.State, 1)
	rxCh := types.RXChannels{StateCh: stateMsgCh}
	tcpPort := 8080
	tcpMsgCh := make(chan types.NetworkMessage, 200)
	isMasterUpdate := make(chan bool)

	tcpPort = runTCPServerAndClient(rxCh, tcpMsgCh, peerUpdateCh, tcpPort)
	go runUDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)

	// The example message. We just send one of these every second.
	go func() {
		stateMsg := types.State{ID:"hah"}
		for {
			//helloMsg.Iter++
			tcpmsg := types.NetworkMessage{stateMsg, types.All, "statemsg"}
			tcpMsgCh <- tcpmsg
			time.Sleep(1 * time.Second)
		}
	}()
	fmt.Println("Started")
	for {
		select {
		case a := <-stateMsgCh:
			fmt.Println("Got TCP message: ", a)
			//case a := <-helloRx:
			//fmt.Printf("Received: %#v\n", a)
		}
	}
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
