package types

import (
	"fmt"
	"strings"
)

const N_FLOORS = 4
const N_BUTTONS = 3

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
	Orders    []ButtonEvent
}

type OrderMatrix [N_FLOORS][N_BUTTONS]bool

type GlobalOrderMap map[string]OrderMatrix

type Receipient int

const (
	All Receipient = iota
	Master
)

/* #### Structs ####*/

type ElevState struct {
	ID          string
	Behavior    Behavior `json:"behavior"`
	Floor       int      `json:"floor"`
	Direction   Dir      `json:"direction"`
	Available   bool
	Obstruction bool
}

type NetworkMessage struct {
	Data       interface{}
	Receipient Receipient
	ChAddr     string
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
	for b := 0; b < N_BUTTONS; b++ {
		i2s := map[int]string{0: "up", 1: "down", 2: "cab"}
		s = append(s, fmt.Sprintf(i2s[b]+"\t"))
		for f := 0; f < N_FLOORS; f++ {
			b2i := map[bool]int8{false: 0, true: 1}
			s = append(s, fmt.Sprintf("%d   ", b2i[mat[f][b]]))
		}
		s = append(s, "\n")
	}
	return strings.Join(s, "")
}

func (m GlobalOrderMap) String() string {
	var s []string
	for name, mat := range m {
		s = append(s, fmt.Sprintln(name+":"))
		s = append(s, fmt.Sprint(mat))
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
