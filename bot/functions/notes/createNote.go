package notes

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"time"
)

var (
	selectorNoteAddFiles = &tb.ReplyMarkup{}
	btnNoteFilesStop     = selectorNoteAddFiles.Data("✅ Далее", "btnNoteFilesNext")

	selectorNoteIsPublic = &tb.ReplyMarkup{}
	btnNotePublic        = selectorNoteIsPublic.Data("🌐 Общедоступная", "btnNotePublic")
	btnNotePrivate       = selectorNoteIsPublic.Data("🔒 Приватная", "btnNotePrivate")
)

func CreateNote(c tb.Context) error {

	userId := c.Sender().ID
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

		if err != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	newNote := models.Note{
		TravelId:     user.CurrentTravelId,
		OwnerId:      userId,
		OwnerName:    user.Name,
		Name:         "",
		Text:         "",
		Files:        nil,
		CreationTime: time.Now(),
	}
	database.DB.Create(&newNote)

	user.CurrentNoteCreatingId = newNote.Id
	database.DB.Where("tg_id=?", user.TgId).Save(&user)

	return InputNoteName(c)
}

func InputNoteName(c tb.Context) error {
	userId := c.Sender().ID
	states.Input.Mx.RLock()
	states.Input.Map[userId] = states.WaitingForNoteName
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите название заметки")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func InputNoteText(c tb.Context) error {
	userId := c.Sender().ID

	states.Input.Mx.RLock()
	states.Input.Map[userId] = states.WaitingForNoteText
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите текст заметки")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func InputNoteFiles(c tb.Context) error {
	userId := c.Sender().ID

	states.Input.Mx.RLock()
	states.Input.Map[userId] = states.WaitingForNoteFiles
	states.Input.Mx.RUnlock()

	selectorNoteAddFiles.Inline(
		selectorNoteAddFiles.Row(btnNoteFilesStop),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"📎 Если вы хотите прикрепить файлы, отправьте их в чат по одному!",
		selectorNoteAddFiles,
	)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

func EndFilesAdding(c tb.Context) error {
	userId := c.Sender().ID

	states.Input.Mx.RLock()
	delete(states.Input.Map, userId)
	states.Input.Mx.RUnlock()

	selectorNoteIsPublic.Inline(
		selectorNoteIsPublic.Row(btnNotePublic, btnNotePrivate),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите приватность заметки", selectorNoteIsPublic)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func NoteIsPublic(c tb.Context) error {
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

		if errSent != nil {
			log.Println(errSent)
		}
		return functions.InputName(c)
	}

	database.DB.Where("id=?", user.CurrentNoteCreatingId).First(&note)
	if note.Id != 0 {
		note.IsPublic = true
		database.DB.Where("id=?", note.Id).Save(&note)

		sentMsg, errSent := c.Bot().Send(c.Chat(),
			fmt.Sprintf("Заметка %s успешно создана", note.Name))

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if err != nil {
			log.Println(errSent)
		}

		user.CurrentNoteCreatingId = 0
		database.DB.Where("tg_id=?", user.TgId).Save(&user)

		return Menu(c)

	} else {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Заметка не найдена")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if err != nil {
			log.Println(errSent)
		}

		return Menu(c)
	}
}

func NoteIsPrivate(c tb.Context) error {
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

	database.DB.Where("id=?", user.CurrentNoteCreatingId).First(&note)
	if note.Id != 0 {
		note.IsPublic = false
		database.DB.Where("id=?", note.Id).Save(&note)

		sentMsg, errSent := c.Bot().Send(c.Chat(),
			fmt.Sprintf(`Заметка "%s" успешно создана`, note.Name))

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if err != nil {
			log.Println(errSent)
		}

		user.CurrentNoteCreatingId = 0
		database.DB.Where("tg_id=?", user.TgId).Save(&user)

		return Menu(c)

	} else {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Заметка не найдена")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if err != nil {
			log.Println(errSent)
		}

		return Menu(c)
	}
}
