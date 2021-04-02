package peers

import (
	"fmt"
	"net"
	"sort"
	"time"

	"../conn"
)

type Peer struct {
	Id       string
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

	isMasterByte := byte(1)
	isMaster := true

	msg := []byte(id)
	msg = append(msg, isMasterByte)

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		case isMaster = <-isMasterUpdate:
			if isMaster {
				isMasterByte = 1
			} else {
				isMasterByte = 0
			}
			msg[len(msg)-1] = isMasterByte
		}
		if enable {
			conn.WriteTo(msg, addr)
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
		id := ""
		isMaster := false
		if n > 1 {
			id = string(buf[:n-1])
			if buf[n-1] == 1 {
				isMaster = true
			}
		}

		// Adding new connection
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.Id = id
				p.IsMaster = isMaster
				p.lastSeen = time.Now()
				updated = true
				lastSeen[id] = p
			} else {
				//TODO: Determine if map should hold pointer to struct so value can be changed directly
				p = lastSeen[id]
				p.lastSeen = time.Now()
				// Check if master status has changed
				if lastSeen[id].IsMaster != isMaster {
					p.IsMaster = isMaster
					updated = true
				}
				lastSeen[id] = p
			}
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
