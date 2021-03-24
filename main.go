package main

import (
	"fmt"
	"sync"

	"./controller_fsm"
	"./hardware_io"
)

func main() {
	const N_FLOORS = 4

	var wg sync.WaitGroup
	wg.Add(1)

	fmt.Println("### Starting Elevator ###")
	go hardware_io.StartElevatorHardware(N_FLOORS)
	go controller_fsm.StartElevatorController()
	// go elevio.PollButtons()

	wg.Wait()
}
