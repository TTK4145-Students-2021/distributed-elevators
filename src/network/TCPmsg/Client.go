package TCPmsg

import (
	"encoding/json"
	"fmt"
	"net"

	"../../types"
	"../masterselect"
	"../peers"
)

type peerConnection struct {
	peer       peers.Peer
	msgChannel chan Request
}

func ClientHandler(id string, rxChannels types.RXChannels, networkMessage <-chan types.NetworkMessage, pCh <-chan peers.PeerUpdate, isMaster chan<- bool, peerLostCh chan<- string) {
	connectedPeers := map[string]peerConnection{}
	tcpConLostCh := make(chan peers.Peer)
	currentMasterId := id
	for {
		select {
		case pUpdate := <-pCh:
			//fmt.Printf("TCPClient: Got peer update\n")
			//Create array of connectedPeers from map
			previousPeersArray := make([]peers.Peer, 0)
			for _, p := range connectedPeers {
				previousPeersArray = append(previousPeersArray, p.peer)
			}

			//Find new and lost peers compared to last iteration
			newPeers := difference(pUpdate.Peers, previousPeersArray)
			lostPeers := difference(previousPeersArray, pUpdate.Peers)

			//Add new peers, remove lost peers
			for _, p := range newPeers {
				fmt.Println("TCP: New peer:", p)
				msgCh := make(chan Request, 100)
				connectedPeers[p.Id] = peerConnection{p, msgCh}
				go handlePeerConnection(p, msgCh, tcpConLostCh)
			}
			for _, p := range lostPeers {
				fmt.Println("TCP: Lost peer:", p)
				delete(connectedPeers, p.Id)
				peerLostCh <- p.Id
			}
			//Determine if we are master, or should stop being master

			currentMasterId = masterselect.DetermineMaster(id, currentMasterId, pUpdate.Peers, isMaster)
		case pLost := <-tcpConLostCh:
			//fmt.Println("TCP: Lost peer:", pLost)
			//Delete connection if TCP con closes
			delete(connectedPeers, pLost.Id)
			peerLostCh <- pLost.Id
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
					//Send messages to yourself locally, not through tcp
					if p.peer.Id != id {
						p.msgChannel <- req
					} else {
						go HandleMessage(req, rxChannels)
					}
				}
			case types.Master:
				for _, p := range connectedPeers {
					if p.peer.Id == currentMasterId {
						//Send messages to yourself locally, not through tcp
						if p.peer.Id != id {
							p.msgChannel <- req
						} else {
							go HandleMessage(req, rxChannels)
						}
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
		_, err = conn.Write(bytes)
		_, err2 := conn.Write([]byte("\n"))
		if err != nil {
			println("TCP Write to server failed:", err.Error())
			return
		}
		if err2 != nil {
			println("TCP Write to server failed:", err2.Error())
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
