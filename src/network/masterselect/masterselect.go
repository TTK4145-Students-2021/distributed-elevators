package masterselect

import (
	"fmt"
	"sort"
	"strconv"

	"../peers"
)

func DetermineMaster(id string, currentMasterId string, connectedPeers []peers.Peer, isMaster chan<- bool) string {
	//Sort all peers, signal if we are lowest id
	//fmt.Println("peers detrmining", connectedPeers)
	var peers []int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: This elevator id is not a int, reboot with proper integer id")
	}
	//peers := make([]string, len(connectedPeers))
	for _, p := range connectedPeers {
		pInt, _ := strconv.Atoi(p.Id)
		peers = append(peers, pInt)
	}
	sort.Ints(peers)
	fmt.Println("Sorted peers: ", peers)
	fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0])

	if peers[0] == idInt {
		isMaster <- true
	} else {
		//fmt.Println("Removed as master")
		isMaster <- false
	}
	currentMasterId = strconv.Itoa(peers[0])
	return currentMasterId

}
