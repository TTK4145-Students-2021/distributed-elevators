package orders

import (
	"../hardware"
	. "../types"
)

type OrderChannels struct {
	LocalOrderCh 	 	chan OrderMatrix,
	LocalLightCh 	 	chan OrderMatrix,
	ClearedFloorCh     	chan int,
	OrdersFromMasterCh 	chan GlobalOrderMap,
	OrderCopyRequestCh 	chan bool,
	ToMasterCh         	chan NetworkMessage,
	KeyPressCh       	chan ButtonEvent,
}

func StartOrderModule(
	ID string,
	ch.OrderChannels,
) {

	orderList := make(GlobalOrderMap)

	for {
		select {

		case button := <-ch.KeyPressCh:
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

		case floor := <-ch.ClearedFloor:
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

		case orderList = <-ch.OrdersFromMaster:
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

		case <-ch.OrderCopyRequest:
			orderCopy := NetworkMessage{
				Data:       orderList,
				Receipient: Master,
				ChAddr:     "ordercopyresponsech",
			}
			toMaster <- orderCopy
		}
	}
}
