package types

import (
	"fmt"
	"strings"
)

/* Variables */
const N_FLOORS = 4
const N_BUTTONS = 3
const ID = "heis_01"

/* #### Types #### */
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

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type OrderEvent struct {
	ElevID    string
	Completed bool
	Order     ButtonEvent
}

type OrderMatrix [N_FLOORS][N_BUTTONS]bool

type GlobalOrderMap map[string]OrderMatrix

/* #### Structs ####*/

type State struct { //ElevState??
	ID        string
	Behavior  Behavior `json:"behavior"`
	Floor     int      `json:"floor"`
	Direction Dir      `json:"direction"`
	Available bool
}

type Receipient int

const (
	All Receipient = iota
	Master
)

type NetworkMessage struct {
	Data       interface{}
	Receipient Receipient
	ChAddr     string
}
type RXChannels struct {
	StateUpdateCh       chan State          `addr:"stateupdatech"`
	RegisterOrderCh     chan OrderEvent     `addr:"registerorderch"`
	OrdersFromMasterCh  chan GlobalOrderMap `addr:"ordersfrommasterch"`
	OrderCopyRequestCh  chan bool           `addr:"ordercopyrequestch"`
	OrderCopyResponseCh chan GlobalOrderMap `addr:"ordercopyresponsech"`
}

/* #### Basic member functions #### */

func (d Dir) String() string {
	a := []string{"up", "down"}
	return a[int(d)]
}

func (b Behavior) String() string {
	a := []string{"idle", "moving", "doorOpen"}
	return a[int(b)]
}

func (mat OrderMatrix) String() string {
	var s []string
	for i := range mat {
		s = append(s, fmt.Sprintf("f%d: ", i+1))
		for _, n := range mat[i] {
			b2i := map[bool]int8{false: 0, true: 1}
			s = append(s, fmt.Sprintf("%d   ", b2i[n]))
		}
		s = append(s, "\n")
	}
	return strings.Join(s, "")
}

func (mat OrderMatrix) OrderOnFloor(floor int) bool {
	for _, btn := range mat[floor] {
		if btn {
			return true
		}
	}
	return false
}

func (m GlobalOrderMap) String() string {
	var s []string
	for name, mat := range m {
		// fmt.Printf("teller")
		s = append(s, fmt.Sprintln(name+":"))
		s = append(s, fmt.Sprint(mat))
	}
	return strings.Join(s, "")
}
