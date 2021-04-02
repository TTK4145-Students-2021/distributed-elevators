package peers

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"time"

	"../conn"
	"../localip"
)

type Peer struct {
	Id       string
	Ip       string
	IsMaster bool
	lastSeen time.Time
}
type PeerUpdate struct {
	Peers []Peer
	//New   Peer
	//Lost  []Peer
}

const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, isMasterUpdate <-chan bool, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}

	isMaster := true

	msgPeer := Peer{id, localIP, isMaster, time.Now()}
	jsonMsg, _ := json.Marshal(msgPeer)

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		case isMaster = <-isMasterUpdate:
			if isMaster {
				msgPeer.IsMaster = true
			} else {
				msgPeer.IsMaster = false
			}
			jsonMsg, _ = json.Marshal(msgPeer)
		}
		if enable {
			conn.WriteTo(jsonMsg, addr)
		}

	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p Peer
	var pUpdate PeerUpdate
	lastSeen := make(map[string]Peer)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])
		err := json.Unmarshal(buf[:n], &p)
		// Adding new connection, check if new peer or master status has changed
		if err == nil {
			p.lastSeen = time.Now()
			if _, idExists := lastSeen[p.Id]; !idExists {
				updated = true
			} else if lastSeen[p.Id].IsMaster != p.IsMaster {
				updated = true
			}
			lastSeen[p.Id] = p
		}
		// Removing dead connection
		for k, v := range lastSeen {
			if time.Since(v.lastSeen) > timeout {
				updated = true
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			pUpdate.Peers = make([]Peer, 0, len(lastSeen))

			for _, v := range lastSeen {
				pUpdate.Peers = append(pUpdate.Peers, v)
			}
			sort.Slice(pUpdate.Peers, func(i, j int) bool {
				return pUpdate.Peers[i].Id > pUpdate.Peers[j].Id
			})
			peerUpdateCh <- pUpdate
		}
	}
}
