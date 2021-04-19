package orders

import (
	"../hardware_io"
	. "../types"
)

func StartOrderModule(localUpdatedOrders chan<- OrderMatrix, localUpdatedLights chan<- OrderMatrix, registerOrder chan<- OrderEvent, globalUpdatedOrders <-chan GlobalOrderMap, completedOrder <-chan int, orderMergeCh chan<- GlobalOrderMap, requestClientOrderCopy <-chan bool) {

	// testMat := OrderMatrix{}
	// // testMat[1][1] = true
	// // testMat[2][1] = true
	// // testMat[3][2] = true

	globalOrderMap := make(GlobalOrderMap)
	// globalOrderMap[ID] = testMat

	keyPress := make(chan ButtonEvent)
	go hardware_io.PollButtons(keyPress)

	for {
		select {

		case button := <-keyPress:
			/* simple case used for testing new orders direct with FSM*/

			newOrder := OrderEvent{
				ElevID:    ID,
				Completed: false,
				Order:     button}

			registerOrder <- newOrder

		case floor := <-completedOrder:

			for i := 0; i < N_BUTTONS; i++ {

				order := ButtonEvent{
					Floor:  floor,
					Button: ButtonType(i),
				}

				completed := OrderEvent{
					ElevID:    ID,
					Completed: true,
					Order:     order}
				registerOrder <- completed
			}
			println("exiting completed")

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

		case <-requestClientOrderCopy:
			orderMergeCh <- globalOrderMap
		}
	}
}
