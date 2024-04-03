package travel

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

// I hate import cycles, so middleware will be here

func IsTravelOwnerMiddleware(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		var travelData models.Travel
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

		database.DB.Where("id=?", user.CurrentTravelId).First(&travelData)
		if travelData.Id == 0 {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Путешествие не найдено")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			return errSent
		}
		if user.TgId != travelData.OwnerId {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Отказано в доступе\n"+
					"Вы не являетесь создателем путешествия")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return OpenTravelMenu(c)
		}

		return next(c)
	}
}
