package main

import (
	"flag"
	"fmt"
	"time"

	TCPmsg "../messaging"
	"../peers"
)

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

	tcpPort = network.runTCPServerAndClient(rxCh, tcpMsgCh, peerUpdateCh, tcpPort)
	go network.runUDPServer(id, tcpPort, isMasterUpdate, peerUpdateCh)

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
