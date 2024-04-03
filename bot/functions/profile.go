package functions

import (
	"backend-qushedo/database"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorChangeProfile = &tb.ReplyMarkup{}
	btnChangeProfile      = selectorChangeProfile.Data("📝 Изменить мой профиль", "changeProfile")
	btnBackProfile        = selectorChangeProfile.Data("< Назад", "back")
)

func SetupMyProfile(b *tb.Bot) {
	b.Handle(&btnChangeProfile, changeProfile)
	b.Handle(&btnBackProfile, Back)
}

func MyProfile(c tb.Context) error {
	var genderEmoji string

	selectorChangeProfile.Inline(
		selectorChangeProfile.Row(btnChangeProfile),
		selectorChangeProfile.Row(btnBackProfile),
	)

	userId := c.Sender().ID
	user, err := database.GetUserHard(userId)
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
	ageName := DetermineAgeName(user.Age)

	switch user.Male {
	case true:
		genderEmoji = "👨"
	case false:
		genderEmoji = "👩"
	}

	profileInfo := fmt.Sprintf(
		"%s %s, %s \n\n"+
			"📝 О себе:\n%s \n\n"+
			"📍 Геолокация:\n%s",
		genderEmoji, user.Name, ageName, user.Bio, user.Address,
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		profileInfo, selectorChangeProfile)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

func changeProfile(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
	return InputName(c)
}

func DetermineAgeName(age uint) string {
	if age%10 == 1 && age%100 != 11 {
		return fmt.Sprintf("%d год", age)
	} else if age%10 >= 2 && age%10 <= 4 && (age%100 < 10 || age%100 >= 20) {
		return fmt.Sprintf("%d года", age)
	}
	return fmt.Sprintf("%d лет", age)
} // The best feature in the whole project
