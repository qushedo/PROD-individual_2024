package handlers

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/functions/notes"
	"backend-qushedo/models"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

func OnMedia(c tb.Context) error {
	userId := c.Sender().ID
	state := states.Input.Map[userId]

	media := c.Message().Media()
	fileID := media.MediaFile().FileID

	user, err := database.GetUserHard(userId)
	if err != nil {
		err = c.Send("Пользователь не найден")
		if err != nil {
			log.Println(err)
		}
		return functions.InputName(c)
	}

	switch state {
	case states.WaitingForNoteFiles:
		var note models.Note

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		database.DB.Where("id=?", user.CurrentNoteCreatingId).First(&note)
		if note.Id != 0 {
			if fileID != "" {
				note.Files = append(note.Files, models.NoteTag{FileId: fileID})
				database.DB.Where("id=?", note.Id).Save(&note)

				sentMsg, errSent := c.Bot().Reply(c.Message(),
					"Файл успешно добавлен")

				stateSent := states.Sent.Map[c.Sender().ID]
				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return notes.InputNoteFiles(c)

			} else {
				sentMsg, errSent := c.Bot().Reply(c.Message(),
					"Ошибка, отправьте файл")

				stateSent := states.Sent.Map[c.Sender().ID]
				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return notes.InputNoteFiles(c)
			}

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Заметка не найдена")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return notes.Menu(c)
		}

	default:
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Неизвестная команда")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}
}
