package models

import (
	"gorm.io/datatypes"
	"time"
)

type NoteTag struct {
	FileId string
}

type Note struct {
	Id           uint   `json:"id"`
	TravelId     uint   `json:"travel_id"`
	OwnerId      int64  `json:"owner_id"`
	OwnerName    string `json:"owner_name"`
	Name         string `json:"name"`
	Text         string `json:"text"`
	Files        datatypes.JSONSlice[NoteTag]
	IsPublic     bool      `json:"is_public"`
	CreationTime time.Time `json:"creation_time"`
}
