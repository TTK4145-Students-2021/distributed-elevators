package main

import (
	"./controller_fsm"
	"./hardware_io"
	"./orders"
	// "./test"
	"fmt"
)

func main() {
	const N_FLOORS = 4
	localOrdersCh := make(chan hardware_io.ButtonEvent)

	fmt.Println("### Starting Elevator ###")
	go controller_fsm.StartElevatorController(localOrdersCh)
	go orders.StartOrderModule(localOrdersCh)
	// go test.StartElevatorHardware(N_FLOORS)
	for {
	}
}
