package Controller_fsm

import (
	"../hardware_io"
	"fmt"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type State int

const (
	ST_Idle State = iota
	ST_Moving
	ST_DoorOpen
)

type Elevator struct {
	Floor     int
	Direction hardware_io.MotorDirection
	State     State
	Orders    [N_FLOORS][N_BUTTONS]bool
	Lights    [N_FLOORS][N_BUTTONS]bool
}

func StartElevatorController(localOrdersCh <-chan hardware_io.ButtonEvent) {
	println("# Starting Controller FSM #")
	hardware_io.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)
	// orderCompleteCh := make(chan int)

	go hardware_io.PollFloorSensor(floorSensorCh)

	var elevator Elevator

	for {
		select {
		/*
			case floor <- floorSensorCh:
				new_state, new_direction, order_done = onFloor(floor, elevator)

				elevator.State = new_state
				elevator.Floor = floor
				elevator.Direction = new_direction

				if new_state == ST_DoorOpen {
					hardware_io.SetDoorOpenLamp(1)
					startTimer()
				}
				hardware_io.SetMotorDirection(elevator.Direction)
				hardware_io.SetFloorIndicator(elevator.Floor)

				/* 	TODO:
				Send state change to master elevator
				Send confirmed order to Order module
		*/

		case <-floorSensorCh:
			// new_state = on_floor_get_state(floor, elevator)
			new_state := ST_Idle

			switch new_state {
			case ST_Idle:
				println("State Idle")
				elevator.Direction = hardware_io.MD_Stop
				break
			case ST_Moving:
				break
			case ST_DoorOpen:
				elevator.Direction = hardware_io.MD_Stop
				hardware_io.SetDoorOpenLamp(true)
				/* 	TODO:
				startTimer()
				Send state change to master elevator
				Send confirmed order to Order module
				*/
			}
			hardware_io.SetMotorDirection(elevator.Direction)
			hardware_io.SetFloorIndicator(elevator.Floor)

		case in := <-localOrdersCh:
			/* simple case used for testing new orders direct*/
			fmt.Printf("%+v\n", in)
			elevator.Orders[in.Floor][in.Button] = true
		}
	}
}

// func onFloor(floor int, elevator Elevator) {
// 	if order_in_direction(floor) {
// 		return St_DoorOpen,
// 	}

// }

// func on_floor_get_state(floor int, elevator Elevator) {
// 	if order_in_direction(floor, elevator.Direction) {
// 		return ST_DoorOpen
// 	} else if order_above(floor) {
// 		return ST_Moving
// 	} else {
// 		return ST_Idle
// 	}
// }
// func order_in_direction(floor int, direction hardware_io.MotorDirection, order Orders) {
// 	return
// }
