package network

import (
	"encoding/json"
	"fmt"

	"../../types"
	"../kcp"
	"../masterselect"
	"../peers"
)

type peerConnection struct {
	peer       peers.Peer
	msgChannel chan RequestMsg
}

func runTcpClient(id string, rxChannels RXChannels, networkMessage <-chan types.NetworkMessage, pCh chan peers.PeerUpdate, isMaster chan<- bool, peerLostCh chan<- string) {
	connectedPeers := map[string]peerConnection{}
	tcpConLostCh := make(chan peers.Peer)
	currentMasterId := id
	var previousPUpdate peers.PeerUpdate
	for {
		select {
		case pUpdate := <-pCh:
			//If a tcp -
			if pUpdate.TCPconnUpdate {
				pUpdate = previousPUpdate
			}
			//fmt.Printf("TCPClient: Got peer update\n")
			//Create array of connectedPeers from map
			previousPeersArray := make([]peers.Peer, 0)
			for _, p := range connectedPeers {
				previousPeersArray = append(previousPeersArray, p.peer)
			}

			//Find new and lost peers compared to last iteration
			newPeers := getPeerDifference(pUpdate.Peers, previousPeersArray)
			lostPeers := getPeerDifference(previousPeersArray, pUpdate.Peers)

			for _, p := range newPeers {
				//If id is ourself, messages are directly sent to local server, noe need for TCP connection
				if p.Id == id {
					connectedPeers[p.Id] = peerConnection{peer: p}
				} else {
					fmt.Println("TCP: New peer:", p)
					msgCh := make(chan RequestMsg, 100)
					connectedPeers[p.Id] = peerConnection{p, msgCh}
					go handlePeerConnection(p, msgCh, tcpConLostCh)
				}

			}
			for _, p := range lostPeers {
				fmt.Println("TCP: Lost peer:", p)
				delete(connectedPeers, p.Id)
				peerLostCh <- p.Id
			}

			//Determine if we are master, or should stop being master. Signal master module on channel
			currentMasterId = masterselect.DetermineMaster(id, currentMasterId, pUpdate.Peers, isMaster)
		case pLost := <-tcpConLostCh:
			//fmt.Println("TCP: Lost peer:", pLost)
			//Delete connection if TCP con closes
			delete(connectedPeers, pLost.Id)
			peerLostCh <- pLost.Id
			tcpPeerUpdate := peers.PeerUpdate{TCPconnUpdate: true}
			pCh <- tcpPeerUpdate
		case message := <-networkMessage:
			dat, _ := json.Marshal(message.Data)
			req := RequestMsg{
				ElevatorId:    "102", //Add elevator id here
				ChannelAdress: message.ChAddr,
				Data:          dat,
			}
			switch message.Receipient {
			case types.All:
				noUDPcon := len(connectedPeers) == 0
				if noUDPcon {
					go PassMsgOnRxChannel(req, rxChannels)
				}
				for _, p := range connectedPeers {
					//Send messages to yourself locally, not through tcp
					if p.peer.Id != id {
						p.msgChannel <- req
					} else {
						go PassMsgOnRxChannel(req, rxChannels)
					}
				}
			case types.Master:
				noUDPcon := len(connectedPeers) == 0
				if noUDPcon {
					go PassMsgOnRxChannel(req, rxChannels)
				}
				for _, p := range connectedPeers {
					if p.peer.Id == currentMasterId {
						//Send messages to yourself locally, not through tcp
						if p.peer.Id != id {
							p.msgChannel <- req
						} else {
							go PassMsgOnRxChannel(req, rxChannels)
						}
					}
				}
			}
		}

	}
}

func handlePeerConnection(p peers.Peer, msg <-chan RequestMsg, pLostCh chan<- peers.Peer) {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.TcpPort)
	conn, err := kcp.Dial(addr)
	defer func() {
		pLostCh <- p
		/*Connection is not currently being closed if the peer is removed from currentPeers, while TCP has not closed.
		The connection will wait for TCP timeout to close, this should not be a problem*/
		conn.Close()
	}()
	if err != nil {
		fmt.Println("TCP network connection error: ")
		fmt.Println(err)
		return
	}
	for {
		message := <-msg
		bytes, _ := json.Marshal(message)
		//fmt.Println("Sent on network")
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

func getPeerDifference(a, b []peers.Peer) (diff []peers.Peer) {
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
