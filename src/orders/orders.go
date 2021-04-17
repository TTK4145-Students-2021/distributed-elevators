package orders

import (
	"../hardware_io"
	. "../types"
)

func StartOrderModule(localUpdatedOrders chan<- OrderMatrix, localUpdatedLights chan<- OrderMatrix, registerOrder chan<- OrderEvent, globalUpdatedOrders <-chan GlobalOrderMap) {
	globalOrderMap := GlobalOrderMap{}

	keyPress := make(chan ButtonEvent)
	go hardware_io.PollButtons(keyPress)

	for {
		select {

		case button := <-keyPress:
			/* simple case used for testing new orders direct with FSM*/

			newOrder := OrderEvent{ID, button}
			registerOrder <- newOrder

		case globalOrderMap = <-globalUpdatedOrders:
			localOrderMat := globalOrderMap[ID]
			localUpdatedOrders <- localOrderMat

			var localLightsMat OrderMatrix = localOrderMat
			for _, orderMat := range globalOrderMap {
				for i := 0; i < N_FLOORS; i++ {
					for j := 0; j < N_BUTTONS-1; j++ {
						localLightsMat[i][j] = localOrderMat[i][j] || orderMat[i][j]
					}
				}
			}
			localUpdatedLights <- localLightsMat
		}
	}
}
