package notes

import (
	"backend-qushedo/database"
	"backend-qushedo/models"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
)

func ChangePrivacy(c tb.Context, id string) error {
	var (
		note       models.Note
		noteStatus string
		rows       []tb.Row
	)

	noteIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}

	noteId := uint(noteIdInt)

	database.DB.Where("id=?", noteId).Find(&note)
	if note.Id == 0 {
		return c.Send("Заметка не найдена")
	}
	if note.OwnerId == c.Sender().ID {
		note.IsPublic = !note.IsPublic
		database.DB.Where("id=?", note.Id).Save(&note)

		switch note.IsPublic {
		case true:
			noteStatus = "🌐 Общедоступная"
		case false:
			noteStatus = "🔒 Приватная"
		}

		btnPrivacy := selectorNote.Data(noteStatus, "privacy", fmt.Sprintf("privacy_%d", note.Id))
		rows = append(rows, selectorNote.Row(btnPrivacy))
		rows = append(rows, selectorNote.Row(btnDeleteNote))
		rows = append(rows, selectorNote.Row(btnBackNote))
		selectorNote.Inline(rows...)

		_, err = c.Bot().EditReplyMarkup(c.Message(), selectorNote)
		return err
	} else {
		return nil
	}
}
