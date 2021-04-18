package main

import (
	"flag"
	"fmt"
	"time"

	TCPmsg "./networkpkg/messaging"
	peers "./networkpkg/peers"
)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	peerUpdateCh := make(chan peers.PeerUpdate)
	helloMsgCh := make(chan TCPmsg.HelloMsg, 1)
	rxCh := TCPmsg.RXChannels{HelloMsgCh: helloMsgCh}
	tcpPort := 8080
	tcpMsgCh := make(chan TCPmsg.NetworkMessage, 200)
	isMasterUpdate := make(chan bool)

	tcpPort = runTCPServerAndClient(rxCh, tcpMsgCh, peerUpdateCh, tcpPort)
	go runUDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			tcpmsg := TCPmsg.NetworkMessage{helloMsg, TCPmsg.All, "hellomsg"}
			tcpMsgCh <- tcpmsg
			time.Sleep(1 * time.Second)
		}
	}()
	fmt.Println("Started")
	for {
		select {
		case a := <-helloMsgCh:
			fmt.Println("Got TCP message: ", a)
			//case a := <-helloRx:
			//fmt.Printf("Received: %#v\n", a)
		}
	}
}

func runTCPServerAndClient(rxCh TCPmsg.RXChannels, tcpMsgCh <-chan TCPmsg.NetworkMessage, peerUpdateCh <-chan peers.PeerUpdate, tcpPort int) int {

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
