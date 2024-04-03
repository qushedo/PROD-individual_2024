package models

import (
	"gorm.io/datatypes"
	"time"
)

type Debt struct {
	ParticipantId int64
	Amount        int
}

type Transaction struct {
	Id           uint      `json:"id"`
	TravelId     uint      `json:"travel_id"`
	OwnerId      int64     `json:"owner_id"`
	CreatedAt    time.Time `json:"created_at"`
	Participants datatypes.JSONSlice[Debt]
}
