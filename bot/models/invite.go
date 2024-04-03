package models

import "time"

type Invite struct {
	Id            uint      `json:"id"`
	TravelId      uint      `json:"travel_id"`
	Data          string    `json:"data"`
	TravelOwnerId int64     `json:"travel_owner_id"`
	CreationTime  time.Time `json:"creation_time"`
}
