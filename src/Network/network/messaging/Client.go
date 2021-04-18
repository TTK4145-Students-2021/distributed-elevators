package TCPmsg

import (
	"encoding/json"
	"fmt"
	"net"

	"../peers"
)

type TXChannels struct {
	/*StateCh chan //elev.state
	OrderUpdateCH chan //OrderUpdateCH
	AllOrdersCH chan //orders */
	TestCh1 chan TestMSG `addr:"testch1"`
	TestCh2 chan TestMSG `addr:"testch2"`
}

type Client struct {
	txChannels TXChannels
}
type peerConnection struct {
	peer       peers.Peer
	msgChannel chan Request
}

/*type peerConnections struct {
	peers []peerConnection
}*/

func ClientHandler(txCh TXChannels, pCh <-chan peers.PeerUpdate) {
	connectedPeers := map[string]peerConnection{}
	peerLostCh := make(chan peers.Peer)
	for {
		select {
		case pUpdate := <-pCh:
			//Create array of connectedPeers from map
			conPeersArray := make([]peers.Peer, 0)
			for _, p := range connectedPeers {
				conPeersArray = append(conPeersArray, p.peer)
			}
			//Find new and lost peers compared to last iteration
			newPeers := difference(pUpdate.Peers, conPeersArray)
			lostPeers := difference(conPeersArray, pUpdate.Peers)
			//Add new peers, remove lost peers
			if newPeers != nil {
				for _, p := range newPeers {
					msgCh := make(chan Request, 100)
					connectedPeers[p.Id] = peerConnection{p, msgCh}
					go handlePeerConnection(p, msgCh, peerLostCh)
				}
				for _, p := range lostPeers {
					delete(connectedPeers, p.Id)
				}
			}
		case pLost := <-peerLostCh:
			//Delete connection if TCP con closes
			if _, ok := connectedPeers[pLost.Id]; ok {
				delete(connectedPeers, pLost.Id)
			}
		}

	}
}

func handlePeerConnection(p peers.Peer, msg <-chan Request, pLostCh chan<- peers.Peer) {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Network ")
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
		select {
		case message := <-msg:
			bytes, err := json.Marshal(message)
			_, err = conn.Write(bytes)
			_, err = conn.Write([]byte("\n"))
			if err != nil {
				println("Write to server failed:", err.Error())
			}

			println("write to server = ", string(bytes))

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
