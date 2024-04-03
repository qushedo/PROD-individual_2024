package travel

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
)

var (
	selectorTravelsList = &tb.ReplyMarkup{}
	btnCreateNewTravel  = selectorTravelsList.Data("🌄 Новое путешествие", "createTravel")
	btnBackTravels      = selectorTravelsList.Data("< Назад", "btnBackTravels")

	selectorTravel = &tb.ReplyMarkup{}
	BtnNotes       = selectorTravel.Data("📝 Заметки", "travelNotes")
	BtnBuildRoute  = selectorTravel.Data("🗺 Построить маршрут", "buildRoute")
	BtnWeather     = selectorTravel.Data("☀️ Что по погоде?", "weatherInfo")
	BtnPoi         = selectorTravel.Data("🏰 Что посетить?", "btnAttractions")
	BtnRestaurants = selectorTravel.Data("🍴 Где поесть?", "btnRestaurants")
	BtnTickets     = selectorTravel.Data("🎟 Билеты", "btnTickets")
	BtnHotels      = selectorTravel.Data("🏨 Отели", "btnHotels")
	BtnSplitWise   = selectorTravel.Data("💸 Учёт трат", "btnSplitWise")
	btnEditTravel  = selectorTravel.Data("✏️ Редактировать", "btnEditTravel")
	btnBackTravel  = selectorTravelsList.Data("< Назад", "btnBackTravel")

	// Have I already said that I love inlines?
)

func MyTravels(c tb.Context) error {
	var (
		travelsOwner []models.Travel

		travelMember []models.TravelMember
		travel       models.Travel

		rows []tb.Row
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

	user.CurrentTravelId = 0
	database.DB.Where("tg_id=?", user.TgId).Save(&user)
	userId := c.Sender().ID
	database.DB.Where("owner_id=?", userId).Find(&travelsOwner)
	for _, travelOwner := range travelsOwner {
		if travelOwner.Name != "" && travelOwner.Description != "" {
			btnTravel := selectorTravelsList.Data(fmt.Sprintf("%s - Создатель путешествия", travelOwner.Name), "travel", fmt.Sprintf("travel_%d", travelOwner.Id))
			rows = append(rows, selectorTravelsList.Row(btnTravel))

		} else {
			database.DB.Where("id=?", travelOwner.Id).Delete(&travelOwner)
		}
	}
	database.DB.Where("tg_id=?", userId).Find(&travelMember)
	for _, member := range travelMember {
		database.DB.Where("id=?", member.TravelId).First(&travel)
		if travel.Id != 0 {
			btnTravel := selectorTravelsList.Data(fmt.Sprintf("%s - Участник путешествия", travel.Name), "travel", fmt.Sprintf("travel_%d", travel.Id))
			rows = append(rows, selectorTravelsList.Row(btnTravel))
		} else {
			database.DB.Where("travel_id=?", member.TravelId).Delete(&member)
		}
	}

	// TODO It needs to be redone in a good way, but I don't think I'll make it in time

	rows = append(rows, selectorTravelsList.Row(btnCreateNewTravel))
	rows = append(rows, selectorTravelsList.Row(btnBackTravels))

	selectorTravelsList.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Список ваших путешествий", selectorTravelsList)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return errSent
}

func Menu(c tb.Context, id string) error {
	var (
		travel models.Travel
		rows   []tb.Row
	)
	_ = c.Delete()

	stateSent := states.Sent.Map[c.Sender().ID]
	stateSent.Delete(c)

	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent = states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	travelIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	travelId := uint(travelIdInt)

	user.CurrentTravelId = travelId
	database.DB.Where("tg_id=?", user.TgId).Save(&user)
	database.DB.Where("id=?", travelId).Find(&travel)
	if travel.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Путешествие не найдено")

		stateSent = states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	rows = append(rows, selectorTravel.Row(BtnNotes, BtnBuildRoute))
	rows = append(rows, selectorTravel.Row(BtnWeather, BtnPoi))
	rows = append(rows, selectorTravel.Row(BtnRestaurants, BtnHotels))
	rows = append(rows, selectorTravel.Row(BtnTickets, BtnSplitWise))
	if travel.OwnerId == user.TgId {
		rows = append(rows, selectorTravelsList.Row(selectorTravel.Data("✏️ Редактировать", "btnEditTravel")))
	} else {
		rows = append(rows, selectorTravelsList.Row(selectorTravel.Data("📋 Информация", "btnEditTravel")))
	}
	rows = append(rows, selectorTravel.Row(btnBackTravel))

	selectorTravel.Inline(rows...)

	travelDesc := fmt.Sprintf(
		"🏞 %s\n\n"+
			"📖 Описание:\n%s",
		travel.Name, travel.Description,
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		travelDesc, selectorTravel)

	stateSent = states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

func OpenTravelMenu(c tb.Context) error {
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

	return Menu(c, strconv.Itoa(int(user.CurrentTravelId)))
}
