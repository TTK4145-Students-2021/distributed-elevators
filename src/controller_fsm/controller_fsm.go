package controller_fsm

import (
	"fmt"
	"time"

	hw "../hardware_io"
	. "../types"
)

type Elevator struct {
	State  State
	Orders OrderMatrix
	Lights OrderMatrix
}

/*
Orders // Lights
1	|	UP	Down	Cab
2	| 	UP	Down	Cab
3	|	UP	Down	Cab
4	|	UP	Down	Cab
*/
func StartElevatorController(localUpdatedOrders <-chan OrderMatrix, localUpdatedLights <-chan OrderMatrix, updateElevState chan<- State, completedOrder chan<- int) {
	println("# Starting Controller FSM #")
	hw.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)
	door_open := make(chan bool, 5)

	/* init goroutines */
	go hw.PollFloorSensor(floorSensorCh)

	/* init variables */
	e := Elevator{
		State:  State{ID, BH_Idle, -1, DIR_Down, true},
		Orders: OrderMatrix{},
		Lights: OrderMatrix{},
	}
	hw.SetMotorDirection(hw.MD_Down)

	door_close := time.NewTimer(3 * time.Second)
	door_close.Stop()

	for {
		select {

		case e.State.Floor = <-floorSensorCh:
			fmt.Println("FSM: Arrived at floor", e.State.Floor, e.State.Direction)

			switch e.State.Behavior {
			case BH_Idle, BH_DoorOpen:
				hw.SetMotorDirection(hw.MD_Stop)

			case BH_Moving:
				if e.shouldTakeOrder() {
					door_open <- true
					break
				}
				if e.ordersEmpty() {
					hw.SetMotorDirection(hw.MD_Stop)
					e.State.Behavior = BH_Idle
					break
				}
				switch e.State.Direction {
				case DIR_Up:
					if !e.ordersAbove() {
						e.State.Direction = DIR_Down
						hw.SetMotorDirection(hw.MD_Down)
					}
				case DIR_Down:
					if !e.ordersBelow() {
						e.State.Direction = DIR_Up
						hw.SetMotorDirection(hw.MD_Up)
					}
				}
			}

			updateElevState <- e.State

		case <-door_open:
			println("FSM: Door Open")
			e.State.Behavior = BH_DoorOpen
			hw.SetMotorDirection(hw.MD_Stop)
			hw.SetDoorOpenLamp(true)
			door_close.Reset(3 * time.Second)

			completedOrder <- e.State.Floor

		case <-door_close.C:
			println("FSM: Door Closing")
			hw.SetDoorOpenLamp(false)

			if e.ordersEmpty() {
				e.State.Behavior = BH_Idle
				break
			}

			e.State.Direction = e.chooseDirection()
			e.State.Behavior = BH_Moving
			hw.SetMotorDirection(hw.MotorDirection(e.State.Direction))

		case orderMat := <-localUpdatedOrders:
			/* simple case used for testing new orders direct*/
			e.Orders = orderMat
			fmt.Println("FSM: got order")
			fmt.Println(orderMat)
			if e.ordersEmpty() {
				break
			}
			// fmt.Println(orderMat)

			switch e.State.Behavior {
			case BH_Moving:
				break

			case BH_Idle:
				if orderOnFloor(orderMat, e.State.Floor) {
					door_open <- true
					break
				}
				e.State.Direction = e.chooseDirection()
				e.State.Behavior = BH_Moving
				hw.SetMotorDirection(hw.MotorDirection(e.State.Direction))
			case BH_DoorOpen:
				if orderOnFloor(orderMat, e.State.Floor) {
					door_open <- true
					break
				}
			}
			// hw.SetButtonLamp(in.Button, in.Floor, true)

			// case lights
		case lightMat := <-localUpdatedLights:
			for floor, arr := range lightMat {
				for btn, value := range arr {
					hw.SetButtonLamp(ButtonType(btn), floor, value)
				}
			}
			e.Lights = lightMat
		}
	}
}

func (e Elevator) shouldTakeOrder() bool {
	switch e.State.Direction {
	case DIR_Up:
		return e.Orders[e.State.Floor][BT_HallUp] ||
			e.Orders[e.State.Floor][BT_Cab] ||
			(!e.ordersAbove())
	case DIR_Down:
		return e.Orders[e.State.Floor][BT_HallDown] ||
			e.Orders[e.State.Floor][BT_Cab] ||
			(!e.ordersBelow() && e.Orders[e.State.Floor][BT_HallUp])
	}
	return false
}

/*temp*/
func clearOrder(e *Elevator) {
	e.Orders[e.State.Floor][BT_HallUp] = false
	e.Orders[e.State.Floor][BT_HallDown] = false
	e.Orders[e.State.Floor][BT_Cab] = false
}
func (e Elevator) clearLights() {
	for btn := 0; btn < N_BUTTONS; btn++ {
		hw.SetButtonLamp(ButtonType(btn), e.State.Floor, false)
	}
}

func (e Elevator) printOrders() {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			println("floor ", floor, "button ", btn, "value ", e.Orders[floor][btn])
		}
	}
}

func onFloorGetNewBehavior(e Elevator) (Behavior, bool) {

	if e.Orders == [N_FLOORS][N_BUTTONS]bool{} {
		return BH_Idle, false
	} else if e.shouldTakeOrder() {
		return BH_DoorOpen, true
	} else {
		return BH_Moving, false
	}
}

func (e Elevator) chooseDirection() Dir {
	switch e.State.Direction {
	case DIR_Up:
		if e.ordersAbove() {
			return DIR_Up
		} else if e.ordersBelow() {
			return DIR_Down
		} else {
			println("FSM: Fatal error")
		}
	case DIR_Down:
		if e.ordersBelow() {
			return DIR_Down
		} else if e.ordersAbove() {
			return DIR_Up
		} else {
			println("FSM: Fatal error, direction down. Orders above= ", e.ordersAbove())
		}
	}
	return e.State.Direction //MUST REMOVE THIS
}

func (e Elevator) ordersEmpty() bool {
	return e.Orders == OrderMatrix{}
}

func (e Elevator) ordersAbove() bool {
	for floor := e.State.Floor + 1; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func (e Elevator) ordersBelow() bool {
	for floor := 0; floor < e.State.Floor; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func orderOnFloor(mat OrderMatrix, floor int) bool {
	for _, btn := range mat[floor] {
		if btn {
			return true
		}
	}
	return false
}
