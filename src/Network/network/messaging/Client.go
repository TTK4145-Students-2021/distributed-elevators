package TCPmsg

import (
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

func ClientHandler(txCh TXChannels, pCh <-chan peers.PeerUpdate) {
	//connectedPeers := make([]peers.Peer, 0)
	connectedPeers := map[string]peers.Peer{}
	for {
		conPeersArray := make([]peers.Peer, 0)
		for _, p := range connectedPeers {
			conPeersArray = append(conPeersArray, p)
		}
		pUpdate := <-pCh
		newPeers := difference(pUpdate.Peers, conPeersArray)
		lostPeers := difference(conPeersArray, pUpdate.Peers)
		if newPeers != nil {
			for _, p := range newPeers {
				connectedPeers[p.Id] = p
				go addConnection(p)
			}
			for _, p := range newPeers {
				delete(connectedPeers, p)
			}
		}

	}
}

func addConnection(p peers.Peer) {

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
