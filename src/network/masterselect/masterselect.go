package masterselect

import (
	"fmt"
	"sort"
	"strconv"

	"../peers"
)

func DetermineIfMaster(id string, connectedPeers []peers.Peer, isMaster chan<- bool) {
	//Sort all peers, signal if we are lowest id
	fmt.Println("peers detrmining", connectedPeers)
	var currentlyMaster bool
	var peers []int
	idInt, _ := strconv.Atoi(id)
	//peers := make([]string, len(connectedPeers))
	fmt.Println("Created peers: ", peers)
	for _, p := range connectedPeers {
		fmt.Println("Added peer ", p.Id)
		pInt, _ := strconv.Atoi(p.Id)
		peers = append(peers, pInt)
		if p.Id == id {
			currentlyMaster = p.IsMaster
		}
	}
	fmt.Println("Unsorted peers: ", peers)
	sort.Ints(peers)
	fmt.Println("Sorted peers: ", peers)
	fmt.Printf("Elevator %s: Master is elevator %d\n", id, peers[0])

	if peers[0] == idInt && !currentlyMaster {
		isMaster <- true
	} else if peers[0] != idInt && currentlyMaster {
		isMaster <- false
	}

}
