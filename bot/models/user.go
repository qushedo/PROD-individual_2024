package models

type User struct {
	TgId      int64   `json:"tg_id"`
	Name      string  `json:"name"`
	Age       uint    `json:"age"`
	Bio       string  `json:"bio"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Male      bool    `json:"male"` // My answer to feminists

	CurrentTravelId          uint  `json:"current_travel_id"`           // TODO: Do this shit on redis if you have time
	CurrentLocationId        uint  `json:"current_location_id"`         // And this
	CurrentMemberId          int64 `json:"current_member_id"`           // And THIS
	CurrentNoteId            uint  `json:"current_note_id"`             // I won't say anything.
	CurrentNoteCreatingId    uint  `json:"current_note_creating_id"`    // I hate myself for doing this here.
	CurrentTransactionId     uint  `json:"current_transaction_id"`      // OMG
	CurrentTransactionUserId int64 `json:"current_transaction_user_id"` // Jesus christ
}
