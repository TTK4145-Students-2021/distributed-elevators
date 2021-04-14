package types

/* Variables */
const N_FLOORS = 4
const N_BUTTONS = 3

/* Types */
type Behavior int

const (
	BH_Idle Behavior = iota
	BH_Moving
	BH_DoorOpen
)

type Dir int

const (
	DIR_Up = iota
	DIR_Down
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type Elevator struct {
	Behavior  Behavior
	Direction Dir
	Floor     int
	Availeble bool
	Orders    [N_FLOORS][N_BUTTONS]bool
	Lights    [N_FLOORS][N_BUTTONS]bool
}

/* Basic member functions */
// func (d Dir) get_string() string {
// 	a := []string{"DIR_Up", "DIR_Down"}
// 	return a[int(d)]
// }
