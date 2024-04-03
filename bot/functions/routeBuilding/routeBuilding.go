package routeBuilding

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/functions/travel"
	"backend-qushedo/graphhopper"
	"backend-qushedo/models"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorRouteChoose = &tb.ReplyMarkup{}
	btnRouteToStartPos  = selectorRouteChoose.Data("📌 Маршрут до стартовой точки", "btnRouteToStartPos")
	btnRouteTravel      = selectorRouteChoose.Data("🗺 Маршрут путешествия", "btnRouteTravel")
	btnBackRouteChoose  = selectorRouteChoose.Data("< Назад", "btnBackRoute")

	selectorRoute = &tb.ReplyMarkup{}
	btnBackRoute  = selectorRoute.Data("< Назад", "btnBackRoute")
)

func ChooseMode(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}

	selectorRouteChoose.Inline(
		selectorRouteChoose.Row(btnRouteToStartPos),
		selectorRouteChoose.Row(btnRouteTravel),
		selectorRouteChoose.Row(btnBackRoute),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите какой маршрут будем строить", selectorRouteChoose)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func BuildTravelRoute(c tb.Context) error {
	var (
		locations []models.Location
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
	locations = database.GetTravelLocations(user.CurrentTravelId)
	url, err := graphhopper.GetLinkByLocations(locations)
	if err != nil {
		errSend := c.Send(err.Error())
		if errSend != nil {
			log.Println(errSend)
		}
		return travel.OpenTravelMenu(c)
	}
	webApp := &tb.WebApp{URL: url}
	btnRouteOpenWebapp := selectorRoute.WebApp("🗺 Открыть построенный маршрут", webApp)
	selectorRoute.Inline(
		selectorRoute.Row(btnRouteOpenWebapp),
		selectorRoute.Row(btnBackRoute),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Ваш маршрут успешно построен", selectorRoute)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func BuildRouteToStartPos(c tb.Context) error {
	var (
		location models.Location
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
	location = database.GetTravelFirstLocation(user.CurrentTravelId)
	if location.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Для построения маршрута до стартовой точки в путешествии должна быть как минимум 1 локация")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return travel.OpenTravelMenu(c)
	}
	userLocation := graphhopper.GhLocation{
		Address:   user.Address,
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
	}

	startPos := graphhopper.GhLocation{
		Address:   location.Address,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
	}

	url, err := graphhopper.GetLink([]graphhopper.GhLocation{userLocation, startPos})
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			err.Error())

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return travel.OpenTravelMenu(c)
	}
	webApp := &tb.WebApp{URL: url}
	btnRouteOpenWebapp := selectorRoute.WebApp("🗺 Открыть построенный маршрут", webApp)
	selectorRoute.Inline(
		selectorRoute.Row(btnRouteOpenWebapp),
		selectorRoute.Row(btnBackRoute),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Ваш маршрут успешно построен", selectorRoute)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}
