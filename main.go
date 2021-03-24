package main

import (
	"./hardware_io"
	"fmt"
	"sync"
)

const num_floors = 4

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	fmt.Printf("Starting Elevator")
	go hardware_io.StartElevatorHardware(num_floors)

	wg.Wait()
}
