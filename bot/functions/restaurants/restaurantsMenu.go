package restaurants

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"backend-qushedo/yandexRestaurants"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorRestaurants = &tb.ReplyMarkup{}
	btnBackRestaurants  = selectorRestaurants.Data("< Назад", "btnBackRestaurants")
)

func Menu(c tb.Context) error {
	var (
		locations []models.Location
		rows      []tb.Row
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

	database.DB.Order("visit_time_start ASC").Where("travel_id=?", user.CurrentTravelId).Find(&locations)
	for _, location := range locations {
		if !location.VisitTimeStart.IsZero() && !location.VisitTimeEnd.IsZero() && location.Address != "" {
			urlYandexRestaurants := yandexRestaurants.GetLink(location.Address)
			webApp := &tb.WebApp{URL: urlYandexRestaurants}
			btnOpenRestaurants := selectorRestaurants.WebApp(location.Address, webApp)
			rows = append(rows, selectorRestaurants.Row(btnOpenRestaurants))
		} else {
			database.DB.Where("id=?", location.Id).Delete(&location)
		}
	}

	rows = append(rows, selectorRestaurants.Row(btnBackRestaurants))

	selectorRestaurants.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🍴 Выберите локацию", selectorRestaurants)

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
