package orders

import (
	"fmt"

	"../hardware_io"
	. "../types"
)

func StartOrderModule(
	ID string,
	localOrderCh chan<- OrderMatrix,
	localLightCh chan<- OrderMatrix,
	clearedFloor <-chan int,
	toMaster chan<- NetworkMessage,
	ordersFromMaster <-chan GlobalOrderMap,
	orderCopyRequest <-chan bool,
) {

	// testMat := OrderMatrix{}
	// // testMat[1][1] = true
	// // testMat[2][1] = true
	// // testMat[3][2] = true

	orderList := make(GlobalOrderMap)
	// globalOrderMap[ID] = testMat

	keyPress := make(chan ButtonEvent)
	go hardware_io.PollButtons(keyPress)

	for {
		select {

		case button := <-keyPress:
			btn := []ButtonEvent{button}
			newOrder := OrderEvent{
				ElevID:    ID,
				Completed: false,
				Orders:    btn}

			registerNewOrder := NetworkMessage{
				Data:       newOrder,
				Receipient: Master,
				ChAddr:     "registerorderch"}

			toMaster <- registerNewOrder

		case floor := <-clearedFloor:
			orderArray := []ButtonEvent{}
			for btn := 0; btn < N_BUTTONS; btn++ {

				button := ButtonEvent{
					Floor:  floor,
					Button: ButtonType(btn),
				}

				orderArray = append(orderArray, button)
			}
			completedOrder := OrderEvent{
				ElevID:    ID,
				Completed: true,
				Orders:    orderArray}

			registerCompletedOrder := NetworkMessage{
				Data:       completedOrder,
				Receipient: Master,
				ChAddr:     "registerorderch"}
			toMaster <- registerCompletedOrder

		case orderList = <-ordersFromMaster:
			localOrders := orderList[ID]
			localOrderCh <- localOrders

			localLights := localOrders
			for _, orders := range orderList {
				for f := 0; f < N_FLOORS; f++ {
					for b := 0; b < N_BUTTONS-1; b++ {
						localLights[f][b] = localLights[f][b] || orders[f][b]
					}
				}
			}
			localLightCh <- localLights

		case <-orderCopyRequest:
			orderCopy := NetworkMessage{
				Data:       orderList,
				Receipient: Master,
				ChAddr:     "ordercopyresponsech",
			}
			toMaster <- orderCopy
		}
	}
}
