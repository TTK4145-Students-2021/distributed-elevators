package controller_fsm

import (
	"../hardware_io/elevio"
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
	Direction elevio.MotorDirection
	State     State
	Orders    [N_FLOORS][N_BUTTONS]bool
	Lights    [N_FLOORS][N_BUTTONS]bool
}

func StartElevatorController() {
	println("Starting Controller")
	elevio.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)
	// orderCompleteCh := make(chan int)

	go elevio.PollFloorSensor(floorSensorCh)

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
					elevio.SetDoorOpenLamp(1)
					startTimer()
				}
				elevio.SetMotorDirection(elevator.Direction)
				elevio.SetFloorIndicator(elevator.Floor)

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
				elevator.Direction = elevio.MD_Stop
				break
			case ST_Moving:
				break
			case ST_DoorOpen:
				elevator.Direction = elevio.MD_Stop
				elevio.SetDoorOpenLamp(true)
				/* 	TODO:
				startTimer()
				Send state change to master elevator
				Send confirmed order to Order module
				*/
			}
			elevio.SetMotorDirection(elevator.Direction)
			elevio.SetFloorIndicator(elevator.Floor)
		}
		// case <- door_timer:
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
// func order_in_direction(floor int, direction elevio.MotorDirection, order Orders) {
// 	return
// }
