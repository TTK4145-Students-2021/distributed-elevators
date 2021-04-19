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
	TcpPort  int
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

func Transmitter(udpPort int, id string, tcpPort int, isMasterUpdate <-chan bool, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(udpPort)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", udpPort))

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}

	isMaster := true

	msgPeer := Peer{id, localIP, tcpPort, isMaster, time.Now()}
	jsonMsg, _ := json.Marshal(msgPeer)

	for {
		select {
		case <-time.After(interval):
		case isMaster = <-isMasterUpdate:
			fmt.Println("Master update: IsMaster:", isMaster)
			if isMaster {
				msgPeer.IsMaster = true
			} else {
				msgPeer.IsMaster = false
				fmt.Println("Setting master false")
			}
			jsonMsg, _ = json.Marshal(msgPeer)
			fmt.Println("Json bc msg: ", msgPeer)
		}

		conn.WriteTo(jsonMsg, addr)
	}
}

func Receiver(udpPort int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p Peer
	var pUpdate PeerUpdate
	pUpdateTimeout := time.Second
	pUpdateTimer := time.Now()
	lastSeen := make(map[string]Peer)
	conn := conn.DialBroadcastUDP(udpPort)

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

		// Sending update, send at interval to synchronize UDP and TCP connection loss
		if updated || time.Since(pUpdateTimer) > pUpdateTimeout {
			pUpdate.Peers = make([]Peer, 0, len(lastSeen))

			for _, v := range lastSeen {
				pUpdate.Peers = append(pUpdate.Peers, v)
			}
			sort.Slice(pUpdate.Peers, func(i, j int) bool {
				return pUpdate.Peers[i].Id > pUpdate.Peers[j].Id
			})
			peerUpdateCh <- pUpdate
			pUpdateTimer = time.Now()
		}
	}
}
