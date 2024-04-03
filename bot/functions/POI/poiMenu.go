package POI

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/openTripMap"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
)

var (
	selectorLocationsPoi = &tb.ReplyMarkup{}
	btnBackLocationsPoi  = selectorLocationsPoi.Data("< Назад", "btnBackLocationsPoi")

	selectorPoi = &tb.ReplyMarkup{}
	btnBackPoi  = selectorPoi.Data("< Назад", "btnBackPoi")

	selectorPoiInfo = &tb.ReplyMarkup{}
	btnBackPoiInfo  = selectorPoiInfo.Data("< Назад", "btnBackPoiInfo")
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

			btnOpenPoi := selectorLocationsPoi.Data(location.Address, "btnLocationPoi", fmt.Sprintf("locationPoi_%d", location.Id))
			rows = append(rows, selectorLocationsPoi.Row(btnOpenPoi))
		} else {
			database.DB.Where("id=?", location.Id).Delete(&location)
		}
	}

	rows = append(rows, selectorLocationsPoi.Row(btnBackLocationsPoi))

	selectorLocationsPoi.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите локацию", selectorLocationsPoi)

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

func LocationPoi(c tb.Context, id string) error {
	var (
		location models.Location
		rows     []tb.Row
	)
	_ = c.Delete()

	idInt, _ := strconv.Atoi(id)
	locationId := uint(idInt)

	database.DB.Where("id=?", locationId).Find(&location)
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
	user.CurrentLocationId = locationId
	database.DB.Where("tg_id=?", user.TgId).Save(&user)

	features, err := openTripMap.Otm.GetPlaces(openTripMap.Opts{
		Lat:    location.Latitude,
		Long:   location.Longitude,
		Rate:   "3,3h,2,2h",
		Radius: "20000",
	})

	if err != nil {
		log.Println(err)
	}

	limit := 20
	count := 0

	for _, feature := range features.Features {
		if count >= limit {
			break
		}
		btnOpenPoi := selectorPoi.Data(fmt.Sprintf("%d метра - %s", int(feature.Properties.Dist), feature.Properties.Name), "btnPoi", fmt.Sprintf("poi_%s", feature.ID))
		rows = append(rows, selectorPoi.Row(btnOpenPoi))
		count += 1
	}
	rows = append(rows, selectorPoi.Row(btnBackPoi))
	selectorPoi.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Вот список из локаций которые стоит посетить", selectorPoi)

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

func Info(c tb.Context, id string) error {
	var (
		location models.Location
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

	database.DB.Where("id=?", user.CurrentLocationId).Find(&location)

	feature, ok := openTripMap.Otm.GetFeatureByID(id, openTripMap.Opts{
		Lat:    location.Latitude,
		Long:   location.Longitude,
		Rate:   "3,3h,2,2h",
		Radius: "20000",
	})

	selectorPoiInfo.Inline(
		selectorPoiInfo.Row(btnBackPoiInfo),
	)

	if !ok {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Достопримечательность не найдена", selectorPoiInfo)
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

	selectorPoiInfo.Inline(
		selectorPoiInfo.Row(btnBackPoiInfo),
	)

	featureLocation := &tb.Location{
		Lat: float32(feature.Geometry.Coordinates[1]),
		Lng: float32(feature.Geometry.Coordinates[0]),
	}

	sentMsg, errSent := featureLocation.Send(c.Bot(), c.Chat(), &tb.SendOptions{})
	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	featureDesk :=
		fmt.Sprintf("%s", feature.Properties.Name) +
			fmt.Sprintf("\n\n%d метра от вашей лоакции", int(feature.Properties.Dist))

	sentMsg, errSent = c.Bot().Send(c.Chat(),
		featureDesk, selectorPoiInfo)

	stateSent = states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}
	return errSent
}

func BackPoiInfo(c tb.Context) error {
	sentStruct := states.Sent.Map[c.Sender().ID]
	sentStruct.Delete(c)
	return Menu(c)
}
