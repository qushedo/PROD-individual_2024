package functions

import (
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

func Cancel(c tb.Context) error {
	states.Input.Mx.RLock()
	delete(states.Input.Map, c.Sender().ID)
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Reply(c.Message(),
		"Действие отменено")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return MainMenu(c)
}
