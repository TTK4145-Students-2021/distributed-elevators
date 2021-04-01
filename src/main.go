package main

import (
	"fmt"

	"./Orders"
	"./controller_fsm"
	"./hardware_io"
)

func main() {
	const N_FLOORS = 4
	localOrdersCh := make(chan hardware_io.ButtonEvent)

	fmt.Println("### Starting Elevator ###")
	go Controller_fsm.StartElevatorController(localOrdersCh)
	go Orders.StartOrderModule(localOrdersCh)
	// go elevio.PollButtons()

	for {
	}
}
