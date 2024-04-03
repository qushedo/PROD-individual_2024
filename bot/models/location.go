package models

import (
	"time"
)

type Location struct {
	Id             uint      `json:"id"`
	TravelId       uint      `json:"travel_id"`
	TravelOwnerId  int64     `json:"travel_owner_id"`
	Address        string    `json:"address"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	VisitTimeStart time.Time `json:"visit_time_start"`
	VisitTimeEnd   time.Time `json:"visit_time_end"`
}
