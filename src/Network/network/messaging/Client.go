package TCPmsg

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

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
			fmt.Printf("TCPClient: Got peer update\n")
			//Create array of connectedPeers from map
			conPeersArray := make([]peers.Peer, 0)
			for _, p := range connectedPeers {
				conPeersArray = append(conPeersArray, p.peer)
			}
			//Find new and lost peers compared to last iteration
			newPeers := difference(pUpdate.Peers, conPeersArray)
			lostPeers := difference(conPeersArray, pUpdate.Peers)
			fmt.Println("Lost Peers: ", lostPeers)
			//Add new peers, remove lost peers
			for _, p := range newPeers {
				msgCh := make(chan Request, 100)
				connectedPeers[p.Id] = peerConnection{p, msgCh}
				go handlePeerConnection(p, msgCh, peerLostCh)
			}
			for _, p := range lostPeers {
				delete(connectedPeers, p.Id)
			}
		case pLost := <-peerLostCh:
			//Delete connection if TCP con closes
			delete(connectedPeers, pLost.Id)
		case <-time.After(500 * time.Millisecond):
			//This is a temp test, should be removed
			fmt.Println("ConPeers: ", connectedPeers)
			data := &TestMSG{42, "Data boiiii"}
			dat, _ := json.Marshal(data)
			req := Request{
				ElevatorId:    "102",
				ChannelAdress: "testch1",
				Data:          dat,
			}
			for _, p := range connectedPeers {
				p.msgChannel <- req
			}
		}

	}
}

func handlePeerConnection(p peers.Peer, msg <-chan Request, pLostCh chan<- peers.Peer) {
	fmt.Printf("TCPclient: Added peer connection\n")
	addr := fmt.Sprintf("%s:%d", p.Ip, p.TcpPort)
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
			bytes, _ := json.Marshal(message)
			_, _ = conn.Write(bytes)
			_, err = conn.Write([]byte("\n"))
			if err != nil {
				println("Write to server failed:", err.Error())
				return
			}
			//println("write to server = ", string(bytes))
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
