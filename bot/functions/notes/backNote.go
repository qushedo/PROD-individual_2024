package notes

import (
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
)

func BackNote(c tb.Context) error {
	sentStruct := states.Sent.Map[c.Sender().ID]
	sentStruct.Delete(c)
	return Menu(c)
}
