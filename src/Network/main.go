package main

import (
	"fmt"
	TCPmsg "./network/messaging"
	"./network/peers"
)




func main() {
	id := "100"
	peerUpdateCh := make(chan peers.PeerUpdate)
	tCh1 := make(chan TCPmsg.TestMSG, 1)
	rxCh := TCPmsg.RXChannels{TestCh1: tCh1}
	tcpPort := 8080
	txCh := TCPmsg.TXChannels{TestCh1: tCh1}
	isMasterUpdate := make(chan bool)
	go TCPServerAndClient(rxCh, txCh, peerUpdateCh, tcpPort)
	go UDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)
	for{}
}


func TCPServerAndClient(rxCh TCPmsg.RXChannels, txCh TCPmsg.TXChannels, peerUpdateCh <-chan peers.PeerUpdate, tcpPort int) {
	
	portCh := make(chan int, 1)
	server := TCPmsg.NewServer(rxCh)

	//Spawn TCP listen client handler, get assigned port
	go server.ListenAndServe(tcpPort, portCh)
	tcpPort = <-portCh
	fmt.Println("Port ", tcpPort)

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	go TCPmsg.ClientHandler(txCh, peerUpdateCh)
}

func UDPServer(id string, tcpPort int, isMasterUpdate chan bool, peerUpdateCh chan<- peers.PeerUpdate){
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, tcpPort, isMasterUpdate, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
}
