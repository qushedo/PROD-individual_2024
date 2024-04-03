package functions

import (
	"backend-qushedo/database"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorMainMenu     = &tb.ReplyMarkup{}
	btnProfile           = selectorMainMenu.Data("👤 Мой профиль", "myProfile")
	BtnMyTravels         = selectorMainMenu.Data("🌍 Мои путешествия", "myTravels")
	btnLookingForCompany = selectorMainMenu.Data("🔍 Найти попутчиков", "lookingForCompany")
)

func SetupMainMenu(b *tb.Bot) {
	b.Handle(&btnProfile, MyProfile)
}

func MainMenu(c tb.Context) error {
	selectorMainMenu.Inline(
		selectorMainMenu.Row(BtnMyTravels, btnLookingForCompany),
		selectorMainMenu.Row(btnProfile),
	)
	_, err := database.GetUserHard(c.Sender().ID)
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

		return InputName(c)
	}

	stateSent := states.Sent.Map[c.Sender().ID]
	stateSent.Delete(c)
	stateSent = states.Sent.Map[c.Sender().ID]

	msg, err := c.Bot().Send(c.Chat(),
		"🧳 Ваш универсальный помощник в путешествиях.\n"+
			"✅ Все что нужно в путешествии у вас под рукой",
		selectorMainMenu,
	)

	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return err
}

func Back(c tb.Context) error {
	return c.Delete()
}

func MainMenuWithDelete(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
	return MainMenu(c)
}
