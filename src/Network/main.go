package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"./network/bcast"
	"./network/localip"
	TCPmsg "./network/messaging"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	isMasterUpdate := make(chan bool)
	go peers.Transmitter(15647, id, 8080, isMasterUpdate, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	//TCP listener server
	tcpPort := 8080
	portCh := make(chan int, 1)
	tCh1 := make(chan jsonpipe.TestMSG, 1)
	rxch := TCPmsg.RXChannels{TestCh1: tCh1}
	server := TCPmsg.NewServer(rxch)

	//Check if TCP listen port is available, otherwise increment until available port is found

	address := "0.0.0.0:"
	go server.ListenAndServe(address, tcpPort, portCh)
	tcpPort <- portCh
	fmt.Println("Port ", tcpPort)
	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			for _, v := range p.Peers {
				fmt.Printf("  Peer: id:%s, ip: %s, isMaster:%t   \n\n", v.Id, v.Ip, v.IsMaster)

			}
		case a := <-tCh1:
			fmt.Println("Got TCP message: ", a)
			//case a := <-helloRx:
			//fmt.Printf("Received: %#v\n", a)
		}
	}
}

/*func MessageHandler() jsonpipe.Handler {
	return func(response *jsonpipe.Response, request *jsonpipe.Request) {
		fmt.Println("Data: ", request.Data)
		response.Data = "Message received"
	}
}*/

/*OrderHandler() jsonpipe.Handler {
	return func(response *jsonpipe.Response, request *jsonpipe.Request) {
		fmt.Println("Do something with this: ", request.Data)
		response.Data = "Message received"
	}
}*/
