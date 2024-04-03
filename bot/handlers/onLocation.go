package handlers

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/keyboards"
	"backend-qushedo/states"
	"fmt"
	nominatim "github.com/doppiogancio/go-nominatim"
	tb "gopkg.in/telebot.v3"
	"log"
)

func OnLocation(c tb.Context) error {
	userId := c.Sender().ID
	state := states.Input.Map[userId]
	switch state {
	case states.WaitingForGeo:
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		location := c.Message().Location
		address, err := nominatim.ReverseGeocode(float64(location.Lat), float64(location.Lng), "ru")
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Ошибка обмена координат на адрес, попробуйте еще раз")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.ChooseGeo(c)
		}
		user, err := database.GetUser(userId)
		if err != nil {
			log.Println(err)
		}

		user.Address = address.DisplayName
		user.Latitude = float64(location.Lat)
		user.Longitude = float64(location.Lng)
		database.DB.Where("tg_id=?", userId).Save(&user)

		sentMsg, errSent := c.Bot().Send(c.Chat(),
			fmt.Sprintf("Добро пожаловать, %s", user.Name), keyboards.EmptyMenu)

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.MainMenu(c)
	default:
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Неизвестная команда")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent

	}
}
