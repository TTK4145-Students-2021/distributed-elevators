package Orders

import (
	"../hardware_io"
)

func StartOrderModule(localOrdersCh chan<- hardware_io.ButtonEvent) {

	buttonCh := make(chan hardware_io.ButtonEvent)
	go hardware_io.PollButtons(buttonCh)

	for {
		select {

		case button := <-buttonCh:
			/* simple case used for testing new orders direct with FSM*/
			localOrdersCh <- button
		}
	}
}
