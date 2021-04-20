package controller

import (
	"fmt"
	"time"

	hw "../hardware"
	. "../types"
)

type Elevator struct {
	State  ElevState
	orders OrderMatrix
	lights OrderMatrix
}

func StartElevatorController(
	ID string,
	orderUpdateCh <-chan OrderMatrix,
	lightUpdateCh <-chan OrderMatrix,
	clearedFloorCh chan<- int,
	toMasterCh chan<- NetworkMessage,
) {

	println("# Starting Controller FSM #")
	/* init channels */
	floorSensorCh := make(chan int)
	stopSensorCh := make(chan bool)
	obstructionSensorCh := make(chan bool)
	door_openCh := make(chan bool, 200)

	/* init goroutines */
	go hw.PollFloorSensor(floorSensorCh)
	go hw.PollStopButton(stopSensorCh)
	go hw.PollObstructionSwitch(obstructionSensorCh)

	/* init variables */
	e := Elevator{
		State: ElevState{
			ID:        ID,
			Behavior:  BH_Moving,
			Floor:     0,
			Direction: DIR_Down,
			Available: false},
		orders: OrderMatrix{},
		lights: OrderMatrix{},
	}

	doorClose := time.NewTimer(3 * time.Second)
	doorClose.Stop()
	hw.SetMotorDirection(hw.MD_Down)
	errorTimeout := time.NewTimer(5 * time.Second)
	sendState := time.NewTimer(1 * time.Second)

	hw.SetDoorOpenLamp(false)

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
					door_openCh <- true
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

			updateState := NetworkMessage{
				Data:       e.State,
				Receipient: Master,
				ChAddr:     "stateupdatech"}
			toMasterCh <- updateState

		case <-door_openCh:
			println("FSM: Door Open")
			e.State.Behavior = BH_DoorOpen
			hw.SetMotorDirection(hw.MD_Stop)
			hw.SetDoorOpenLamp(true)
			doorClose.Reset(3 * time.Second)
			errorTimeout.Stop()
			clearedFloorCh <- e.State.Floor

		case <-doorClose.C:
			if e.State.Obstruction {
				doorClose.Reset(3 * time.Second)
				break
			}
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
		case newOrders := <-orderUpdateCh:
			e.orders = newOrders
			if e.ordersEmpty() {
				break
			}

			switch e.State.Behavior {
			case BH_Moving:
				break

			case BH_Idle:
				if newOrders.OrderOnFloor(e.State.Floor) {
					door_openCh <- true
					break
				}
				e.State.Direction = e.chooseDirection()
				e.State.Behavior = BH_Moving
				hw.SetMotorDirection(hw.MotorDirection(e.State.Direction))
				errorTimeout.Reset(5 * time.Second)
			case BH_DoorOpen:
				if newOrders.OrderOnFloor(e.State.Floor) {
					door_openCh <- true
					break
				}
			}

		case newLights := <-lightUpdateCh:
			for f, row := range newLights {
				for b, setLamp := range row {
					hw.SetButtonLamp(ButtonType(b), f, setLamp)
				}
			}
			e.lights = newLights

		case <-errorTimeout.C:
			/* Case where elevator gets stuck */
			fmt.Println("FMS: FATAL ERROR - Motor Timout triggered, elevator stuck?")
			e.State.Available = false
			updateState := NetworkMessage{
				Data:       e.State,
				Receipient: Master,
				ChAddr:     "stateupdatech"}
			toMasterCh <- updateState

		case <-sendState.C:
			updateState := NetworkMessage{
				Data:       e.State,
				Receipient: Master,
				ChAddr:     "stateupdatech"}
			sendState.Reset(1 * time.Second)
			toMasterCh <- updateState

		case <-stopSensorCh:
			hw.SetMotorDirection(hw.MD_Stop)

		case obstructed := <-obstructionSensorCh:
			fmt.Println("Got obstruction, value: ", obstructed)
			if obstructed {
				e.State.Obstruction = true
				e.State.Available = false
			} else {
				e.State.Obstruction = false
				e.State.Available = true
			}
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

func (e Elevator) chooseDirection() Dir {
	switch e.State.Direction {
	case DIR_Up:
		if e.ordersAbove() {
			return DIR_Up
		} else if e.ordersBelow() {
			return DIR_Down
		} else {
			println("FSM: Fatal error")
			return DIR_Down
		}
	case DIR_Down:
		if e.ordersBelow() {
			return DIR_Down
		} else if e.ordersAbove() {
			return DIR_Up
		} else {
			println("FSM: Fatal error")
			return DIR_Up
		}
	}
	return e.State.Direction
}

func (e Elevator) ordersEmpty() bool {
	return e.orders == OrderMatrix{}
}

func (e Elevator) ordersAbove() bool {
	for f := e.State.Floor + 1; f < N_FLOORS; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			if e.orders[f][b] {
				return true
			}
		}
	}
	return false
}

func (e Elevator) ordersBelow() bool {
	for f := 0; f < e.State.Floor; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			if e.orders[f][b] {
				return true
			}
		}
	}
	return false
}