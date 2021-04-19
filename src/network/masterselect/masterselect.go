package masterselect

import (
	"fmt"
	"sort"
	"strconv"

	"../peers"
)

func DetermineMaster(id string, currentMasterId string, connectedPeers []peers.Peer, isMaster chan<- bool) string {
	//Sort all peers, signal if we are lowest id
	fmt.Println("peers detrmining", connectedPeers)
	var peers []int
	currentlyMaster := id == currentMasterId
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: This elevator id is not a int, reboot with proper integer id")
	}
	//peers := make([]string, len(connectedPeers))
	fmt.Println("Created peers: ", peers)
	for _, p := range connectedPeers {
		fmt.Println("Added peer ", p.Id)
		pInt, _ := strconv.Atoi(p.Id)
		peers = append(peers, pInt)
	}
	sort.Ints(peers)
	fmt.Println("Sorted peers id: ", peers)
	fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0])

	if peers[0] == idInt && !currentlyMaster {
		isMaster <- true
	} else if peers[0] != idInt && currentlyMaster {
		fmt.Println("Removed as master")
		isMaster <- false
	}
	currentMasterId = strconv.Itoa(peers[0])
	return currentMasterId

}
