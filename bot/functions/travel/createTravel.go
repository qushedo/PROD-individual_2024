package travel

import (
	"backend-qushedo/database"
	"backend-qushedo/models"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

func NewTravel(c tb.Context) error {
	database.DB.Create(&models.Travel{
		OwnerId: c.Sender().ID,
	})

	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
	return InputTravelName(c)
}

func InputTravelName(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelName
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите название путешествия")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func InputTravelDesc(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelDescription
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите описание путешествия")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}
