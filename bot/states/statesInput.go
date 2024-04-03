package states

import "sync"

// TODO: Do this shit on redis if you have time
//  probably didn't have time for this

type inputMap struct {
	Mx  sync.RWMutex
	Map map[int64]int
}

var Input *inputMap

const (
	WaitingForName = 1
	WaitingForAge  = 2
	WaitingForBio  = 3
	WaitingForGeo  = 4

	WaitingForTravelName        = 5
	WaitingForTravelDescription = 6

	WaitingForTravelNameEdit = 7
	WaitingForTravelDescEdit = 8

	WaitingForTravelLocation       = 9
	WaitingForTravelVisitTimeStart = 10
	WaitingForTravelVisitTimeEnd   = 11

	WaitingForNoteName  = 12
	WaitingForNoteText  = 13
	WaitingForNoteFiles = 14

	WaitingForTransactionAmount = 15
)

func MakeInputMap() {
	Input = &inputMap{Map: make(map[int64]int), Mx: sync.RWMutex{}}
}
