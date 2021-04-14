package controller_fsm

import (
	"../hardware_io"
	"fmt"
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type Behavior int

const (
	BH_Idle Behavior = iota
	BH_Moving
	BH_DoorOpen
)

type Dir int

const (
	DIR_Up = iota
	DIR_Down
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type Elevator struct {
	behavior  Behavior
	direction Dir
	floor     int
	available bool
	orders    [N_FLOORS][N_BUTTONS]bool
	lights    [N_FLOORS][N_BUTTONS]bool
}

/*
Orders // Lights
1	|	UP	Down	Cab
2	| 	UP	Down	Cab
3	|	UP	Down	Cab
4	|	UP	Down	Cab
*/
func StartElevatorController(localOrdersCh <-chan hardware_io.ButtonEvent) {
	println("# Starting Controller FSM #")
	hardware_io.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)

	/* init goroutines */
	go hardware_io.PollFloorSensor(floorSensorCh)

	/* init variables */
	elevator := Elevator{
		behavior:  BH_Idle,
		direction: DIR_Down,
		floor:     -1,
		available: true,
		orders:    [N_FLOORS][N_BUTTONS]bool{},
		lights:    [N_FLOORS][N_BUTTONS]bool{},
	}
	hardware_io.SetMotorDirection(hardware_io.MD_Down)

	door_timed_out := time.NewTimer(3 * time.Second)
	door_timed_out.Stop()

	for {
		select {

		case elevator.floor = <-floorSensorCh:
			fmt.Println("Arrived at floor", elevator.floor+1)

			new_behavior, _ := onFloorGetNewBehavior(elevator)

			switch new_behavior {
			case BH_Moving:
				println("FSM: Moving")
				elevator.direction = chooseDirection(elevator)
				hardware_io.SetMotorDirection(hardware_io.MotorDirection(elevator.direction))

			case BH_DoorOpen:
				println("FSM: Door Open")
				hardware_io.SetMotorDirection(hardware_io.MD_Stop)
				hardware_io.SetDoorOpenLamp(true)
				door_timed_out.Reset(3 * time.Second)

				clearOrder(&elevator)
				clearLights(elevator)
				/* 	TODO:
				Clear order
				Send state change to master elevator
				Send confirmed order to Order module
				*/
			case BH_Idle:
				println("FSM: Idle")
				hardware_io.SetMotorDirection(hardware_io.MD_Stop)
			}
			elevator.behavior = new_behavior

		case <-door_timed_out.C:
			println("Door Timed Out")
			hardware_io.SetDoorOpenLamp(false)

			new_behavior, _ := onFloorGetNewBehavior(elevator)
			switch new_behavior {
			case BH_Moving:
				elevator.direction = chooseDirection(elevator)
			case BH_DoorOpen, BH_Idle:

			}

		case in := <-localOrdersCh:
			/* simple case used for testing new orders direct*/
			elevator.orders[in.Floor][in.Button] = true
			new_direction := chooseDirection(elevator)

			switch elevator.behavior {
			case BH_Idle:
				hardware_io.SetMotorDirection(hardware_io.MotorDirection(new_direction))
			case BH_Moving:
				break
			case BH_DoorOpen:
				hardware_io.SetMotorDirection(hardware_io.MotorDirection(new_direction))
			}
			fmt.Printf("%+v\n", in)
			printOrders(elevator)
			hardware_io.SetButtonLamp(in.Button, in.Floor, true)
		}
	}
}

func orderOnFloor(elevator Elevator) bool {
	switch elevator.direction {
	case DIR_Up:
		return elevator.orders[elevator.floor][BT_HallUp] ||
			elevator.orders[elevator.floor][BT_Cab] ||
			!ordersAbove(elevator)
	case DIR_Down:
		return elevator.orders[elevator.floor][BT_HallDown] ||
			elevator.orders[elevator.floor][BT_Cab] ||
			!ordersBelow(elevator)
	}
	return false
}

/*temp*/
func clearOrder(elevator *Elevator) {
	elevator.orders[elevator.floor][BT_HallUp] = false
	elevator.orders[elevator.floor][BT_HallDown] = false
	elevator.orders[elevator.floor][BT_Cab] = false
}
func clearLights(elevator Elevator) {
	for btn := 0; btn < N_BUTTONS; btn++ {
		hardware_io.SetButtonLamp(hardware_io.ButtonType(btn), elevator.floor, false)
	}
}

func printOrders(elevator Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			println("floor ", floor, "button ", btn, "value ", elevator.orders[floor][btn])
		}
	}
}

func onFloorGetNewBehavior(elevator Elevator) (Behavior, bool) {

	if elevator.orders == [N_FLOORS][N_BUTTONS]bool{} {
		return BH_Idle, false
	} else if orderOnFloor(elevator) {
		return BH_DoorOpen, true
	} else {
		return BH_Moving, false
	}
}

func chooseDirection(elevator Elevator) Dir {
	switch elevator.direction {
	case DIR_Up:
		if ordersAbove(elevator) {
			return DIR_Up
		} else if ordersBelow(elevator) {
			return DIR_Down
		} else {
			println("Fatal error, direction up")
		}
	case DIR_Down:
		if ordersBelow(elevator) {
			return DIR_Down
		} else if ordersAbove(elevator) {
			return DIR_Up
		} else {
			println("Fatal error, direction down. Orders above= ", ordersAbove(elevator))
		}
	}
	return elevator.direction //MUST REMOVE THIS
}

func ordersAbove(elevator Elevator) bool {
	for floor := elevator.floor + 1; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func ordersBelow(elevator Elevator) bool {
	for floor := 0; floor < elevator.floor; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}
