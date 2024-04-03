package tickets

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"backend-qushedo/yandexTravels"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
)

var (
	selectorTicketsTransport = &tb.ReplyMarkup{}
	btnAvia                  = selectorTicketsTransport.Data("✈️ Самолет", "btnAvia")
	btnTrain                 = selectorTicketsTransport.Data("🚂 Поезд", "btnTrain")
	btnBus                   = selectorTicketsTransport.Data("🚌 Автобус", "btnBus")
	btnBackTicketsTransport  = selectorTicketsTransport.Data("< Назад", "btnBackTicketsTransport")

	selectorTicketsTrain = &tb.ReplyMarkup{}
	btnBackTicketsTrain  = selectorTicketsTrain.Data("< Назад", "btnBackTicketsBus")

	selectorTicketsBus = &tb.ReplyMarkup{}
	btnBackTicketsBus  = selectorTicketsBus.Data("< Назад", "btnBackTicketsTrain")

	selectorTicketsAvia = &tb.ReplyMarkup{}
	btnBackTicketsAvia  = selectorTicketsAvia.Data("< Назад", "btnBackTicketsAvia")
)

func ChooseTransport(c tb.Context) error {
	_ = c.Delete()
	selectorTicketsTransport.Inline(
		selectorTicketsTransport.Row(btnAvia),
		selectorTicketsTransport.Row(btnTrain),
		selectorTicketsTransport.Row(btnBus),
		selectorTicketsTransport.Row(btnBackTicketsTransport),
	)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🛄 Выберите тип транспорта", selectorTicketsTransport)

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

func AviaMenu(c tb.Context) error {
	var (
		locations     []models.Location
		travel        models.Travel
		travelOwner   models.User
		travelMembers []models.TravelMember
		rows          []tb.Row
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
	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)
	database.DB.Where("tg_id=?", travel.OwnerId).Find(&travelOwner)
	database.DB.Where("travel_id=?", travel.Id).Find(&travelMembers)

	if len(locations) >= 1 {
		optsFromUser :=
			yandexTravels.GetAviaURLOpts{
				FromLat: user.Latitude,
				FromLng: user.Longitude,
				ToLat:   locations[0].Latitude,
				ToLng:   locations[0].Longitude,
				Adults:  "1",
				When:    locations[0].VisitTimeStart.AddDate(0, 0, -1),
			}
		urlAviaFromUser, errGetAviaLinkFromUser := yandexTravels.GetAviaLink(optsFromUser)
		if errGetAviaLinkFromUser == nil {
			webApp := &tb.WebApp{URL: urlAviaFromUser}
			btnOpenAvia := selectorTicketsAvia.WebApp(fmt.Sprintf("%s - %s | %s", "Ваше местоположение", locations[0].Address, locations[0].VisitTimeStart.AddDate(0, 0, -1).Format("02.01.2006")), webApp)
			rows = append(rows, selectorTicketsTrain.Row(btnOpenAvia))
		}
		for index, locationFrom := range locations {
			if !locationFrom.VisitTimeStart.IsZero() && !locationFrom.VisitTimeEnd.IsZero() && locationFrom.Address != "" {
				if index+1 < len(locations) {
					locationTo := locations[index+1]
					if !locationTo.VisitTimeStart.IsZero() && !locationTo.VisitTimeEnd.IsZero() && locationTo.Address != "" {
						adults, minors := countAdultsAndMinors(travelMembers)
						if travelOwner.Age >= 18 {
							adults++
						} else {
							minors++
						}
						opts :=
							yandexTravels.GetAviaURLOpts{
								FromLat:  locationFrom.Latitude,
								FromLng:  locationFrom.Longitude,
								ToLat:    locationTo.Latitude,
								ToLng:    locationTo.Longitude,
								Adults:   strconv.Itoa(adults),
								Children: strconv.Itoa(minors),
								When:     locationTo.VisitTimeStart,
							}

						urlAvia, errGetAviaLink := yandexTravels.GetAviaLink(opts)
						if errGetAviaLink != nil {
							continue
						}
						webApp := &tb.WebApp{URL: urlAvia}
						btnOpenAvia := selectorTicketsAvia.WebApp(fmt.Sprintf("%s - %s | %s", locationFrom.Address, locationTo.Address, locationFrom.VisitTimeEnd.Format("02.01.2006")), webApp)
						rows = append(rows, selectorTicketsAvia.Row(btnOpenAvia))
					} else {
						database.DB.Where("id=?", locationTo.Id).Delete(&locationTo)
					}
				} else {
					break
				}

			} else {
				database.DB.Where("id=?", locationFrom.Id).Delete(&locationFrom)
			}
		}
	}

	rows = append(rows, selectorTicketsAvia.Row(btnBackTicketsAvia))

	selectorTicketsAvia.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"✈️ Выберите маршрут", selectorTicketsAvia)

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

func TrainMenu(c tb.Context) error {
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

	if len(locations) >= 1 {
		urlTrainFromUser, errGetTrainLinkFromUser := yandexTravels.GetBusLinkByCords(user.Latitude, user.Longitude, locations[0].Latitude, locations[0].Longitude, locations[0].VisitTimeStart.AddDate(0, 0, -1))
		if errGetTrainLinkFromUser == nil {
			webApp := &tb.WebApp{URL: urlTrainFromUser}
			btnOpenTrain := selectorTicketsTrain.WebApp(fmt.Sprintf("%s - %s | %s", "Ваше местоположение", locations[0].Address, locations[0].VisitTimeStart.AddDate(0, 0, -1).Format("02.01.2006")), webApp)
			rows = append(rows, selectorTicketsTrain.Row(btnOpenTrain))
		}

		for index, locationFrom := range locations {
			if !locationFrom.VisitTimeStart.IsZero() && !locationFrom.VisitTimeEnd.IsZero() && locationFrom.Address != "" {
				if index+1 < len(locations) {
					locationTo := locations[index+1]
					if !locationTo.VisitTimeStart.IsZero() && !locationTo.VisitTimeEnd.IsZero() && locationTo.Address != "" {
						urlTrain, errGetTrainLink := yandexTravels.GetTrainLink(locationFrom, locationTo, locationFrom.VisitTimeEnd)
						if errGetTrainLink != nil {
							continue
						}
						webApp := &tb.WebApp{URL: urlTrain}
						btnOpenTrain := selectorTicketsTrain.WebApp(fmt.Sprintf("%s - %s | %s", locationFrom.Address, locationTo.Address, locationFrom.VisitTimeEnd.Format("02.01.2006")), webApp)
						rows = append(rows, selectorTicketsTrain.Row(btnOpenTrain))
					} else {
						database.DB.Where("id=?", locationTo.Id).Delete(&locationTo)
					}
				} else {
					break
				}

			} else {
				database.DB.Where("id=?", locationFrom.Id).Delete(&locationFrom)
			}
		}
	}

	rows = append(rows, selectorTicketsTrain.Row(btnBackTicketsTrain))

	selectorTicketsTrain.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🚂 Выберите маршрут", selectorTicketsTrain)

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

func BusMenu(c tb.Context) error {
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
	if len(locations) >= 1 {
		urlBusFromUser, errGetBusLinkFromUser := yandexTravels.GetBusLinkByCords(user.Latitude, user.Longitude, locations[0].Latitude, locations[0].Longitude, locations[0].VisitTimeStart.AddDate(0, 0, -1))
		if errGetBusLinkFromUser == nil {
			webApp := &tb.WebApp{URL: urlBusFromUser}
			btnOpenBus := selectorTicketsBus.WebApp(fmt.Sprintf("%s - %s | %s", "Ваше местоположение", locations[0].Address, locations[0].VisitTimeStart.AddDate(0, 0, -1).Format("02.01.2006")), webApp)
			rows = append(rows, selectorTicketsBus.Row(btnOpenBus))
		}
		for index, locationFrom := range locations {
			if !locationFrom.VisitTimeStart.IsZero() && !locationFrom.VisitTimeEnd.IsZero() && locationFrom.Address != "" {
				if index+1 < len(locations) {
					locationTo := locations[index+1]
					if !locationTo.VisitTimeStart.IsZero() && !locationTo.VisitTimeEnd.IsZero() && locationTo.Address != "" {
						urlBus, errGetBusLink := yandexTravels.GetBusLink(locationFrom, locationTo, locationFrom.VisitTimeEnd)
						if errGetBusLink != nil {
							continue
						}
						webApp := &tb.WebApp{URL: urlBus}
						btnOpenTrains := selectorTicketsBus.WebApp(fmt.Sprintf("%s - %s | %s", locationFrom.Address, locationTo.Address, locationFrom.VisitTimeEnd.Format("02.01.2006")), webApp)
						rows = append(rows, selectorTicketsBus.Row(btnOpenTrains))
					} else {
						database.DB.Where("id=?", locationTo.Id).Delete(&locationTo)
					}
				} else {
					break
				}

			} else {
				database.DB.Where("id=?", locationFrom.Id).Delete(&locationFrom)
			}
		}
	}

	rows = append(rows, selectorTicketsBus.Row(btnBackTicketsBus))

	selectorTicketsBus.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🚌 Выберите маршрут", selectorTicketsBus)

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

func countAdultsAndMinors(members []models.TravelMember) (int, int) {
	var adults int
	var minors int
	for _, member := range members {
		if member.Age >= 18 {
			adults++
		} else {
			minors++
		}
	}
	return adults, minors
}
