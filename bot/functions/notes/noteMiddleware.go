package notes

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

func IsNoteOwnerMiddleware(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		var noteData models.Note
		user, err := database.GetUserHard(c.Sender().ID)
		if err != nil {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Пользователь не найден")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if err != nil {
				log.Println(errSent)
			}

			return functions.InputName(c)
		}

		database.DB.Where("id=?", user.CurrentNoteId).First(&noteData)
		if noteData.Id == 0 {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Заметка не найдена")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			return errSent
		}
		if user.TgId != noteData.OwnerId {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Отказано в доступе\n"+
					"Вы не являетесь создателем заметки")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if err != nil {
				log.Println(errSent)
			}
			return OpenNoteMenu(c)
		}

		return next(c)
	}
}
