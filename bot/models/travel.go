package models

type Travel struct {
	Id          uint   `json:"id"`
	OwnerId     int64  `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
