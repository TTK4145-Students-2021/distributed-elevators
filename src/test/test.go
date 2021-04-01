package test

import (
	"../hardware_io"
	"fmt"
)

func StartElevatorHardware(numFloors int) {
	fmt.Println("Starting Hardware_IO")
	hardware_io.Init("localhost:15657", numFloors)

	var d hardware_io.MotorDirection = hardware_io.MD_Up
	//hardware_io.SetMotorDirection(d)

	drv_buttons := make(chan hardware_io.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go hardware_io.PollButtons(drv_buttons)
	go hardware_io.PollFloorSensor(drv_floors)
	go hardware_io.PollObstructionSwitch(drv_obstr)
	go hardware_io.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			hardware_io.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = hardware_io.MD_Down
			} else if a == 0 {
				d = hardware_io.MD_Up
			}
			hardware_io.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				hardware_io.SetMotorDirection(hardware_io.MD_Stop)
			} else {
				hardware_io.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := hardware_io.ButtonType(0); b < 3; b++ {
					hardware_io.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
