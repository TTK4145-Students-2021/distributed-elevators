package orders

import (
	"../hardware_io"
	. "../types"
)

func StartOrderModule(localOrdersCh chan<- ButtonEvent, newOrder chan<- OrderEvent) {

	buttonCh := make(chan ButtonEvent)
	go hardware_io.PollButtons(buttonCh)

	for {
		select {

		case button := <-buttonCh:
			/* simple case used for testing new orders direct with FSM*/
			localOrdersCh <- button

			new_order := OrderEvent{ID, button}
			newOrder <- new_order

		}
	}
}
