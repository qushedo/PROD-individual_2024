package notes

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/functions/travel"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
)

var (
	selectorNotes = &tb.ReplyMarkup{}
	btnCreateNote = selectorNotes.Data("📝 Добавить заметку", "btnCreateNote")
	btnBackNotes  = selectorNotes.Data("< Назад", "btnBackNotes")

	selectorNote  = &tb.ReplyMarkup{}
	btnDeleteNote = selectorNote.Data("🗑 Удалить", "btnDeleteNote")
	btnBackNote   = selectorNote.Data("< Назад", "btnBackNote")
)

func Menu(c tb.Context) error {

	var (
		notes []models.Note
		rows  []tb.Row

		travelData models.Travel
	)

	_ = c.Delete()

	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	database.DB.Order("creation_time DESC").Where("travel_id=? AND is_public=TRUE OR owner_id=? ", user.CurrentTravelId, user.TgId).Find(&notes)
	database.DB.Where("id=?", user.CurrentTravelId).First(&travelData)
	if travelData.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Путешествие не найдено")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if err != nil {
			log.Println(errSent)
		}
		return travel.MyTravels(c)
	}

	for _, note := range notes {
		if note.Name != "" && note.Text != "" {
			btnNote := selectorNotes.Data(fmt.Sprintf("%s | %s - %s", note.Name, note.CreationTime.Format("02.01.2006"), note.OwnerName), "note", fmt.Sprintf("note_%d", note.Id))
			rows = append(rows, selectorNotes.Row(btnNote))

		} else {
			database.DB.Where("id=?", note.Id).Delete(&note)
		}
	}

	rows = append(rows, selectorNotes.Row(btnCreateNote))
	rows = append(rows, selectorNotes.Row(btnBackNotes))

	selectorNotes.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf(`Заметки к путешествию "%s"`, travelData.Name), selectorNotes)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func NoteMenu(c tb.Context, id string) error {
	var (
		note       models.Note
		noteStatus string
		rows       []tb.Row
	)

	_ = c.Delete()
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}
	stateSent := states.Sent.Map[user.TgId]

	noteIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	noteId := uint(noteIdInt)
	user.CurrentNoteId = noteId
	database.DB.Where("tg_id=?", user.TgId).Save(&user)
	database.DB.Where("id=?", noteId).Find(&note)
	if note.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Заметка не найдена")

		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	switch note.IsPublic {
	case true:
		noteStatus = "🌐 Общедоступная"
	case false:
		noteStatus = "🔒 Приватная"

	}
	noteContent := fmt.Sprintf(
		"📝 Заметка \"%s\" | %s - %s\n\n"+
			"Текст:\n%s",
		note.Name, note.OwnerName, note.CreationTime.Format("02.01.2006"), note.Text,
	)

	btnPrivacy := selectorNote.Data(noteStatus, "privacy", fmt.Sprintf("privacy_%d", note.Id))
	rows = append(rows, selectorNote.Row(btnPrivacy))
	rows = append(rows, selectorNote.Row(btnDeleteNote))
	rows = append(rows, selectorNote.Row(btnBackNote))

	selectorNote.Inline(rows...)
	if note.Files != nil {
		for _, file := range note.Files {
			fileToSend, err := c.Bot().FileByID(file.FileId)
			if err != nil {
				return err
			}
			msg, err := c.Bot().Send(c.Chat(), &tb.Photo{File: fileToSend})
			if err != nil {
				msg, err = c.Bot().Send(c.Chat(), &tb.Document{File: fileToSend})
				if err != nil {
					return err
				}
			}

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
			states.Sent.Map[user.TgId] = stateSent
			states.Sent.Mx.RUnlock()

		} // Костыль, но вроде окей
	}

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		noteContent, selectorNote)

	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

func OpenNoteMenu(c tb.Context) error {
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	return NoteMenu(c, strconv.Itoa(int(user.CurrentNoteId)))
}
