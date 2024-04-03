package travel

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
	selectorEditTravel     = &tb.ReplyMarkup{}
	btnEditTravelName      = selectorEditTravel.Data("✏️ Название", "btnEditTravelName")
	btnEditTravelDesc      = selectorEditTravel.Data("📄 Описание", "btnEditTravelDesc")
	btnEditTravelMembers   = selectorEditTravel.Data("👥 Участники", "btnEditTravelMembers")
	btnEditTravelLocations = selectorEditTravel.Data("📍 Локации", "btnEditTravelLocations")
	btnDeleteTravel        = selectorEditTravel.Data("🗑 Удалить", "btnDeleteTravel")
	btnBackEditTravel      = selectorEditTravel.Data("< Назад", "btnBackEditTravel")

	selectorDeleteTravel = &tb.ReplyMarkup{}
	btnTravelDeleteYes   = selectorDeleteTravel.Data("✅ Да, удалить", "travelDeleteYes")
	btnTravelDeleteNo    = selectorDeleteTravel.Data("❌ Нет, оставить", "travelDeleteNo")

	// I really love inlines
)

func EditTravel(c tb.Context) error {
	var travel models.Travel
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
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
	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)
	if travel.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Путешествие не найдено")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}
	if user.TgId == travel.OwnerId {
		selectorEditTravel.Inline(
			selectorEditTravel.Row(btnEditTravelName, btnEditTravelDesc),
			selectorEditTravel.Row(btnEditTravelMembers, btnEditTravelLocations),
			selectorEditTravel.Row(btnDeleteTravel),
			selectorEditTravel.Row(btnBackEditTravel),
		)
	} else {
		selectorEditTravel.Inline(
			selectorEditTravel.Row(btnEditTravelMembers, btnEditTravelLocations),
			selectorEditTravel.Row(btnBackEditTravel),
		)
	}

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf(`Редактирование "%s"`, travel.Name), selectorEditTravel)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func EditTravelName(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}

	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelNameEdit
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите новое название путешествия")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func EditTravelDesc(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}

	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelDescEdit
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите новое описание путешествия")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func DeleteTravel(c tb.Context) error {
	var travel models.Travel

	selectorDeleteTravel.Inline(
		selectorDeleteTravel.Row(btnTravelDeleteNo, btnTravelDeleteYes),
	)
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
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

	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Вы уверены, что хотите удалить путешествие\n"+fmt.Sprintf(`"%s?"`, travel.Name), selectorDeleteTravel)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func YesDeleteTravel(c tb.Context) error {
	var travel models.Travel
	var members []models.TravelMember
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

	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf(`Путешествие "%s" успешно удалено`, travel.Name))

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	database.DB.Where("travel_id=?", travel.Id).Delete(&members)
	database.DB.Delete(&travel)

	return MyTravels(c)
}
