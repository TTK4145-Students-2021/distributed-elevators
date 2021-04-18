package TCPmsg

import (
	"encoding/json"
	"fmt"
	"net"

	"../../types"
	"../peers"
)

type peerConnection struct {
	peer       peers.Peer
	msgChannel chan Request
}



func ClientHandler(networkMessage <-chan types.NetworkMessage, pCh <-chan peers.PeerUpdate) {
	connectedPeers := map[string]peerConnection{}
	peerLostCh := make(chan peers.Peer)
	for {
		select {
		case pUpdate := <-pCh:
			//fmt.Printf("TCPClient: Got peer update\n")

			//Create array of connectedPeers from map
			conPeersArray := make([]peers.Peer, 0)
			for _, p := range connectedPeers {
				conPeersArray = append(conPeersArray, p.peer)
			}

			//Find new and lost peers compared to last iteration
			newPeers := difference(pUpdate.Peers, conPeersArray)
			lostPeers := difference(conPeersArray, pUpdate.Peers)

			//Add new peers, remove lost peers
			for _, p := range newPeers {
				fmt.Println("TCP: New peer:", p)
				msgCh := make(chan Request, 100)
				connectedPeers[p.Id] = peerConnection{p, msgCh}
				go handlePeerConnection(p, msgCh, peerLostCh)
			}
			for _, p := range lostPeers {
				fmt.Println("TCP: Lost peer:", p)
				delete(connectedPeers, p.Id)
			}
		case pLost := <-peerLostCh:
			//fmt.Println("TCP: Lost peer:", pLost)
			//Delete connection if TCP con closes
			delete(connectedPeers, pLost.Id)
		case message := <-networkMessage:
			dat, _ := json.Marshal(message.Data)
			req := Request{
				ElevatorId:    "102", //Add elevator id here
				ChannelAdress: message.ChAddr,
				Data:          dat,
			}
			switch message.Receipient {
			case types.All:
				for _, p := range connectedPeers {
					p.msgChannel <- req
				}
			case types.Master:
				for _, p := range connectedPeers {
					if p.peer.IsMaster {
						p.msgChannel <- req
					}
				}
			}
		}

	}
}

func handlePeerConnection(p peers.Peer, msg <-chan Request, pLostCh chan<- peers.Peer) {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.TcpPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("TCP network connection error: ")
		fmt.Println(err)
		return
	}
	defer func() {
		pLostCh <- p
		/*Connection is not currently being closed if the peer is removed from currentPeers.
		The connection will wait for TCP timeout to close*/
		conn.Close()
	}()
	for {
		message := <-msg
		bytes, _ := json.Marshal(message)
		_, _ = conn.Write(bytes)
		_, err = conn.Write([]byte("\n"))
		if err != nil {
			println("Write to server failed:", err.Error())
			return
		}

	}
}

func difference(a, b []peers.Peer) (diff []peers.Peer) {
	m := make(map[string]bool)
	for _, item := range b {
		m[item.Id] = true
	}

	for _, item := range a {
		if _, ok := m[item.Id]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
