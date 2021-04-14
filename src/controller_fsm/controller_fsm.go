package controller_fsm

import (
	"fmt"
	"time"

	hw "../hardware_io"
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
func StartElevatorController(localOrdersCh <-chan hw.ButtonEvent) {
	println("# Starting Controller FSM #")
	hw.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)
	door_open := make(chan bool, 5)

	/* init goroutines */
	go hw.PollFloorSensor(floorSensorCh)

	/* init variables */
	e := Elevator{
		behavior:  BH_Idle,
		direction: DIR_Down,
		floor:     -1,
		available: true,
		orders:    [N_FLOORS][N_BUTTONS]bool{},
		lights:    [N_FLOORS][N_BUTTONS]bool{},
	}
	hw.SetMotorDirection(hw.MD_Down)

	door_close := time.NewTimer(3 * time.Second)
	door_close.Stop()

	for {
		select {

		case e.floor = <-floorSensorCh:
			fmt.Println("Arrived at floor", e.floor, e.direction.get_string())

			switch e.behavior {
			case BH_Idle, BH_DoorOpen:
				hw.SetMotorDirection(hw.MD_Stop)

			case BH_Moving:
				if e.shouldTakeOrder() {
					door_open <- true
					break
				}
				if e.ordersEmpty() {
					hw.SetMotorDirection(hw.MD_Stop)
					e.behavior = BH_Idle
					break
				}
				switch e.direction {
				case DIR_Up:
					if !e.ordersAbove() {
						e.direction = DIR_Down
						hw.SetMotorDirection(hw.MD_Down)
					}
				case DIR_Down:
					if !e.ordersBelow() {
						e.direction = DIR_Up
						hw.SetMotorDirection(hw.MD_Up)
					}
				}
			}

		case <-door_open:
			println("FSM: Door Open")
			e.behavior = BH_DoorOpen
			hw.SetMotorDirection(hw.MD_Stop)
			hw.SetDoorOpenLamp(true)
			door_close.Reset(3 * time.Second)
			clearOrder(&e)
			e.clearLights()

		case <-door_close.C:
			println("Door Closing")
			hw.SetDoorOpenLamp(false)

			if e.ordersEmpty() {
				e.behavior = BH_Idle
				break
			}

			e.direction = e.chooseDirection()
			e.behavior = BH_Moving
			hw.SetMotorDirection(hw.MotorDirection(e.direction))

		case in := <-localOrdersCh:
			/* simple case used for testing new orders direct*/
			e.orders[in.Floor][in.Button] = true

			switch e.behavior {
			case BH_Idle:
				if in.Floor == e.floor {
					door_open <- true
					break
				}
				e.direction = e.chooseDirection()
				e.behavior = BH_Moving
				hw.SetMotorDirection(hw.MotorDirection(e.direction))
			case BH_Moving:
				break
			case BH_DoorOpen:
				if in.Floor == e.floor {
					door_open <- true
				}
			}
			fmt.Printf("%+v\n", in)
			hw.SetButtonLamp(in.Button, in.Floor, true)
		}
	}
}

func (e Elevator) shouldTakeOrder() bool {
	switch e.direction {
	case DIR_Up:
		return e.orders[e.floor][BT_HallUp] ||
			e.orders[e.floor][BT_Cab] ||
			(!e.ordersAbove())
	case DIR_Down:
		return e.orders[e.floor][BT_HallDown] ||
			e.orders[e.floor][BT_Cab] ||
			(!e.ordersBelow() && e.orders[e.floor][BT_HallUp])
	}
	return false
}

/*temp*/
func clearOrder(e *Elevator) {
	e.orders[e.floor][BT_HallUp] = false
	e.orders[e.floor][BT_HallDown] = false
	e.orders[e.floor][BT_Cab] = false
}
func (e Elevator) clearLights() {
	for btn := 0; btn < N_BUTTONS; btn++ {
		hw.SetButtonLamp(hw.ButtonType(btn), e.floor, false)
	}
}

func (elevator Elevator) printOrders() {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			println("floor ", floor, "button ", btn, "value ", elevator.orders[floor][btn])
		}
	}
}

func (d Dir) get_string() string {
	a := []string{"DIR_Up", "DIR_Down"}
	return a[int(d)]
}

func onFloorGetNewBehavior(e Elevator) (Behavior, bool) {

	if e.orders == [N_FLOORS][N_BUTTONS]bool{} {
		return BH_Idle, false
	} else if e.shouldTakeOrder() {
		return BH_DoorOpen, true
	} else {
		return BH_Moving, false
	}
}

func (e Elevator) chooseDirection() Dir {
	switch e.direction {
	case DIR_Up:
		if e.ordersAbove() {
			return DIR_Up
		} else if e.ordersBelow() {
			return DIR_Down
		} else {
			println("Fatal error")
		}
	case DIR_Down:
		if e.ordersBelow() {
			return DIR_Down
		} else if e.ordersAbove() {
			return DIR_Up
		} else {
			println("Fatal error, direction down. Orders above= ", e.ordersAbove())
		}
	}
	return e.direction //MUST REMOVE THIS
}

func (e Elevator) ordersEmpty() bool {
	return e.orders == [N_FLOORS][N_BUTTONS]bool{}
}

func (e Elevator) ordersAbove() bool {
	for floor := e.floor + 1; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func (e Elevator) ordersBelow() bool {
	for floor := 0; floor < e.floor; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}
