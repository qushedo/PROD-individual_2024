package notes

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupNotes(b *tb.Bot) {

	noteGroup := b.Group()

	noteGroup.Handle(&travel.BtnNotes, Menu)
	noteGroup.Handle(&btnCreateNote, CreateNote)
	noteGroup.Handle(&btnNoteFilesStop, EndFilesAdding)
	noteGroup.Handle(&btnNotePublic, NoteIsPublic)
	noteGroup.Handle(&btnNotePrivate, NoteIsPrivate)
	noteGroup.Handle(&btnBackNote, BackNote)
	noteGroup.Handle(&btnBackNotes, travel.OpenTravelMenu)

	noteGroup.Use(IsNoteOwnerMiddleware)
	noteGroup.Handle(&btnDeleteNote, DeleteNote)
	noteGroup.Handle(&btnNoteDeleteYes, DeleteNoteYes)
	noteGroup.Handle(&btnNoteDeleteNo, Menu)
}
