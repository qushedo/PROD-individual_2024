package notes

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorDeleteNote = &tb.ReplyMarkup{}
	btnNoteDeleteNo    = selectorDeleteNote.Data("❌ Нет, оставить", "btnNoteDeleteNo")
	btnNoteDeleteYes   = selectorDeleteNote.Data("✅ Да, удалить", "btnNoteDeleteYes")
)

func DeleteNote(c tb.Context) error {
	var note models.Note

	selectorDeleteNote.Inline(
		selectorDeleteNote.Row(btnNoteDeleteNo, btnNoteDeleteYes),
	)
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		err = c.Send("Пользователь не найден")
		if err != nil {
			log.Println(err)
		}
		return functions.InputName(c)
	}
	sentState := states.Sent.Map[user.TgId]
	sentState.Delete(c)
	database.DB.Where("id=?", user.CurrentNoteId).Find(&note)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Вы уверены, что хотите удалить заметку\n"+fmt.Sprintf(`"%s?"`, note.Name), selectorDeleteNote)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func DeleteNoteYes(c tb.Context) error {
	var note models.Note
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

	database.DB.Where("id=?", user.CurrentNoteId).Find(&note)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf(`Заметка "%s" успешно удалена`, note.Name))

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if err != nil {
		log.Println(errSent)
	}

	database.DB.Delete(&note)

	return Menu(c)
}
