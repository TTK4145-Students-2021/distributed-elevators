package controller_fsm

import (
	"fmt"
	"time"

	hw "../hardware_io"
	. "../types"
)

type Elevator struct {
	State  State
	orders OrderMatrix
	lights OrderMatrix
}

/*
Orders // Lights
1	|	UP	Down	Cab
2	| 	UP	Down	Cab
3	|	UP	Down	Cab
4	|	UP	Down	Cab
*/

/* ISSUES
- What happens if elevator box power turned off ?
-
*/

func StartElevatorController(localUpdatedOrders <-chan OrderMatrix, localUpdatedLights <-chan OrderMatrix, updateElevState chan<- State, completedOrder chan<- int) {
	println("# Starting Controller FSM #")
	hw.Init("localhost:15657", N_FLOORS)

	/* init channels */
	floorSensorCh := make(chan int)
	stopSensorCh := make(chan bool)
	door_open := make(chan bool, 5)

	/* init goroutines */
	go hw.PollFloorSensor(floorSensorCh)
	go hw.PollStopButton(stopSensorCh)

	/* init variables */
	e := Elevator{
		State: State{
			ID:        ID,
			Behavior:  BH_Idle,
			Floor:     -1,
			Direction: DIR_Down,
			Available: true},
		orders: OrderMatrix{},
		lights: OrderMatrix{},
	}

	doorClose := time.NewTimer(3 * time.Second)
	doorClose.Stop()
	hw.SetMotorDirection(hw.MD_Down)
	errorTimeout := time.NewTimer(5 * time.Second)

	for {
		select {

		case e.State.Floor = <-floorSensorCh:
			fmt.Println("FSM: Arrived at floor", e.State.Floor, e.State.Direction)

			switch e.State.Behavior {
			case BH_Idle, BH_DoorOpen:
				hw.SetMotorDirection(hw.MD_Stop)
				errorTimeout.Stop()
			case BH_Moving:
				if e.shouldTakeOrder() {
					door_open <- true
					break
				}
				if e.ordersEmpty() {
					hw.SetMotorDirection(hw.MD_Stop)
					e.State.Behavior = BH_Idle
					errorTimeout.Stop()
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
				errorTimeout.Reset(5 * time.Second)
			}
			e.State.Available = true
			updateElevState <- e.State

		case <-door_open:
			println("FSM: Door Open")
			hw.SetMotorDirection(hw.MD_Stop)
			hw.SetDoorOpenLamp(true)
			doorClose.Reset(3 * time.Second)
			errorTimeout.Stop()
			completedOrder <- e.State.Floor

		case <-doorClose.C:
			println("FSM: Door Closing")
			hw.SetDoorOpenLamp(false)

			if e.ordersEmpty() {
				e.State.Behavior = BH_Idle
				errorTimeout.Stop()
				break
			} else {
				e.State.Direction = e.chooseDirection()
				e.State.Behavior = BH_Moving
				hw.SetMotorDirection(hw.MotorDirection(e.State.Direction))
				errorTimeout.Reset(5 * time.Second)
			}
		case orderMat := <-localUpdatedOrders:
			e.orders = orderMat
			// fmt.Println("FSM: got order")
			// fmt.Println(orderMat)
			if e.ordersEmpty() {
				break
			}

			switch e.State.Behavior {
			case BH_Moving:
				break

			case BH_Idle:
				if orderMat.OrderOnFloor(e.State.Floor) {
					door_open <- true
					break
				}
				e.State.Direction = e.chooseDirection()
				e.State.Behavior = BH_Moving
				hw.SetMotorDirection(hw.MotorDirection(e.State.Direction))
				errorTimeout.Reset(5 * time.Second)
			case BH_DoorOpen:
				if orderMat.OrderOnFloor(e.State.Floor) {
					door_open <- true
					break
				}
			}

		case lightMat := <-localUpdatedLights:
			for floor, arr := range lightMat {
				for btn, setLamp := range arr {
					hw.SetButtonLamp(ButtonType(btn), floor, setLamp)
				}
			}
			e.lights = lightMat

		case <-errorTimeout.C:
			/* Case where elevator gets stuck */
			fmt.Println("FMS: FATAL ERROR - Motor Timout triggered, elevator stuck?")
			e.State.Available = false
			updateElevState <- e.State

		case <-stopSensorCh:
			hw.SetMotorDirection(hw.MD_Stop)
		}
	}
}

func (e Elevator) shouldTakeOrder() bool {
	switch e.State.Direction {
	case DIR_Up:
		return e.orders[e.State.Floor][BT_HallUp] ||
			e.orders[e.State.Floor][BT_Cab] ||
			(!e.ordersAbove())
	case DIR_Down:
		return e.orders[e.State.Floor][BT_HallDown] ||
			e.orders[e.State.Floor][BT_Cab] ||
			(!e.ordersBelow() && e.orders[e.State.Floor][BT_HallUp])
	}
	return false
}

/*temp*/
func clearOrder(e *Elevator) { //REMOVE
	e.orders[e.State.Floor][BT_HallUp] = false
	e.orders[e.State.Floor][BT_HallDown] = false
	e.orders[e.State.Floor][BT_Cab] = false
}
func (e Elevator) clearLights() { //REMOVE
	for btn := 0; btn < N_BUTTONS; btn++ {
		hw.SetButtonLamp(ButtonType(btn), e.State.Floor, false)
	}
}

func (e Elevator) printOrders() { //REMOVE
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			println("floor ", floor, "button ", btn, "value ", e.orders[floor][btn])
		}
	}
}

func onFloorGetNewBehavior(e Elevator) (Behavior, bool) { //REMOVE

	if e.orders == [N_FLOORS][N_BUTTONS]bool{} {
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
	return e.orders == OrderMatrix{}
}

func (e Elevator) ordersAbove() bool {
	for floor := e.State.Floor + 1; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

func (e Elevator) ordersBelow() bool {
	for floor := 0; floor < e.State.Floor; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.orders[floor][btn] {
				return true
			}
		}
	}
	return false
}

// func orderOnFloor(mat OrderMatrix, floor int) bool {
// 	for _, btn := range mat[floor] {
// 		if btn {
// 			return true
// 		}
// 	}
// 	return false
// }
