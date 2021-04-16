package types

import (
	"fmt"
	"strings"
)

/* Variables */
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

type OrderMatrix [N_FLOORS][N_BUTTONS]bool

type GlobalOrderMap map[string]OrderMatrix

/* #### Structs ####*/

type State struct {
	Behavior  Behavior `json:"behavior"`
	Floor     int      `json:"floor"`
	Direction Dir      `json:"direction"`
	Availeble bool
}

/* #### Basic member functions #### */
func (d Dir) String() string {
	a := []string{"DIR_Up", "DIR_Down"}
	return a[int(d)]
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

func (m GlobalOrderMap) String() string {
	var s []string
	for name, mat := range m {
		// fmt.Printf("teller")
		s = append(s, fmt.Sprintln(name+":"))
		s = append(s, fmt.Sprint(mat))
	}
	return strings.Join(s, "")
}