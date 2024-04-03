package models

import "time"

type TravelMember struct {
	Name     string    `json:"name"`
	Age      uint      `json:"age"`
	TgId     int64     `json:"tg_id"`
	TravelId uint      `json:"travel_id"`
	JoinTime time.Time `json:"join_time"`
}
