package states

import (
	tb "gopkg.in/telebot.v3"
)

func AddToSentState(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		stateSent := Sent.Map[c.Sender().ID]
		Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, c.Message().ID)
		Sent.Map[c.Sender().ID] = stateSent
		Sent.Mx.RUnlock()

		return next(c)
	}
}
