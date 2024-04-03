package travel

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	"github.com/doppiogancio/go-nominatim/shared"
	"github.com/zsefvlol/timezonemapper"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
	_ "time/tzdata"
)

var (
	selectorLocationsList = &tb.ReplyMarkup{}
	btnAddNewLocation     = selectorLocationsList.Data("📌 Новая локация", "addLocation")
	btnBackLocations      = selectorLocationsList.Data("< Назад", "btnBackLocations")

	selectorLocationIsCorrect = &tb.ReplyMarkup{}
	btnCorrectLocation        = selectorLocationIsCorrect.Data("✅ Да, всё верно", "locationIsCorrect")
	btnIncorrectLocation      = selectorLocationIsCorrect.Data("❌ Нет, исправить", "locationIsIncorrect")

	selectorLocationMenu = &tb.ReplyMarkup{}
	btnDeleteLocation    = selectorLocationMenu.Data("🗑 Удалить", "btnDeleteLocation")
	btnBackLocation      = selectorLocationMenu.Data("< Назад", "btnBackLocation")

	selectorDeleteLocation = &tb.ReplyMarkup{}
	btnLocationDeleteYes   = selectorDeleteLocation.Data("✅ Да, удалить", "locationDeleteYes")
	btnLocationDeleteNo    = selectorDeleteLocation.Data("❌ Нет, оставить", "locationDeleteNo")

	// I love inlines
)

func LocationsMenu(c tb.Context) error {
	var (
		locations []models.Location
	)
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

	err = c.Delete()
	if err != nil {
		log.Println(err)
	}
	rows := make([]tb.Row, 0)
	database.DB.Order("visit_time_start ASC").Where("travel_id=?", user.CurrentTravelId).Find(&locations)
	for _, location := range locations {
		if !location.VisitTimeStart.IsZero() && !location.VisitTimeEnd.IsZero() && location.Address != "" {
			btnLocation := selectorLocationsList.Data(fmt.Sprintf("%s - %s", location.Address, location.VisitTimeStart.Format("02.01.2006")), "location", fmt.Sprintf("location_%d", location.Id))
			rows = append(rows, selectorLocationsList.Row(btnLocation))
		} else {
			database.DB.Delete(&location)
		}
	}
	rows = append(rows, selectorLocationsList.Row(btnAddNewLocation))
	rows = append(rows, selectorLocationsList.Row(btnBackLocations))

	selectorLocationsList.Inline(rows...)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Список локаций путешествия", selectorLocationsList)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func LocationMenu(c tb.Context, id string) error {
	var location models.Location

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

	locationIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	locationId := uint(locationIdInt)
	user.CurrentLocationId = locationId
	database.DB.Where("tg_id=?", c.Sender().ID).Save(&user)
	database.DB.Where("id=?", locationId).Find(&location)
	if location.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Локация не найдена")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	selectorLocationMenu.Inline(
		selectorLocationMenu.Row(btnDeleteLocation),
		selectorLocationMenu.Row(btnBackLocation),
	)
	timezone := timezonemapper.LatLngToTimezoneString(location.Latitude, location.Longitude)
	timeZoneLoc, errLoad := time.LoadLocation(timezone)
	if errLoad != nil {
		log.Println(errLoad)
	}
	now := time.Now()
	localTime := now.In(timeZoneLoc)

	locationDesc := fmt.Sprintf(
		"📍 Локация: %s\n\n"+
			"🕒 Начало посещения: %s\n"+
			"🕒 Окончание посещения: %s\n"+
			"⏱ Часовой пояс: UTC %s",
		location.Address,
		location.VisitTimeStart.Format("02.01.2006 15:04"),
		location.VisitTimeEnd.Format("02.01.2006 15:04"),
		localTime.Format("Z07:00"),
	)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		locationDesc, selectorLocationMenu)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

//func OpenLocationMenu(c tb.Context) error {
//	user, err := database.GetUserHard(c.Sender().ID)
//	if err != nil {
//		err = c.Send("Пользователь не найден")
//		if err != nil {
//			log.Println(err)
//		}
//		return functions.InputName(c)
//	}
//
//	return Menu(c, strconv.Itoa(int(user.CurrentLocationId)))
//}

func DeleteLocation(c tb.Context) error {
	var location models.Location

	selectorDeleteLocation.Inline(
		selectorDeleteLocation.Row(btnLocationDeleteNo, btnLocationDeleteYes),
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

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("Вы уверены, что хотите удалить локацию\n"+`"%s"?`, fmt.Sprintf("%s - %s", location.Address, location.VisitTimeStart.Format("02.01.2006"))), selectorDeleteLocation)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func YesDeleteLocation(c tb.Context) error {
	var location models.Location
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

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf(`Локация "%s" успешно удалена`, fmt.Sprintf("%s - %s", location.Address, location.VisitTimeStart.Format("02.01.2006"))))

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	database.DB.Delete(&location)

	return LocationsMenu(c)
}

func AddLocation(c tb.Context) error {
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

	location := &models.Location{
		TravelOwnerId: user.TgId,
		TravelId:      user.CurrentTravelId,
	}
	database.DB.Create(&location)

	user.CurrentLocationId = location.Id
	database.DB.Where("tg_id=?", user.TgId).Save(&user)

	err = c.Delete()
	if err != nil {
		log.Println(err)
	}
	return InputTravelLocation(c)
}

func InputTravelLocation(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelLocation
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🔍 Введите название локации для посещения в следующем формате:\n\n"+
			"Город, Страна (Например: Москва, Россия) \n\n"+
			"Если не удаётся найти локацию, используйте расширенный формат:\n\n"+
			"Город, Район, Область, Страна")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}

func IsLocationCorrect(c tb.Context, address shared.Address, strAddress string, coordinate shared.Coordinate) error {
	var location models.Location
	userId := c.Sender().ID
	selectorLocationIsCorrect.Inline(
		selectorLocationIsCorrect.Row(btnCorrectLocation, btnIncorrectLocation),
	)
	user, err := database.GetUser(userId)
	if err != nil {
		log.Println(err)
	}
	database.DB.Where("id = ?", user.CurrentLocationId).First(&location)
	location.Address = strAddress
	location.Latitude = coordinate.Latitude
	location.Longitude = coordinate.Longitude
	database.DB.Where("id=?", location.Id).Save(&location)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("Локация: \n%s\n\nВсё верно?", address.DisplayName), selectorLocationIsCorrect)

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

func LocationCorrect(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		return err
	}

	return InputVisitTimeStart(c)
}

func LocationIncorrect(c tb.Context) error {
	var location models.Location
	err := c.Delete()
	if err != nil {
		return err
	}

	userId := c.Sender().ID
	user, err := database.GetUser(userId)
	if err != nil {
		log.Println(err)
	}

	database.DB.Where("id = ?", user.CurrentLocationId).First(&location)
	location.Address = ""
	database.DB.Where("id=?", location.Id).Save(&location)
	return InputTravelLocation(c)
}

func InputVisitTimeStart(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelVisitTimeStart
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"📅 Введите дату и время посещения локации в следующем формате:\n\n"+
			"ДД.ММ.ГГГГ ЧЧ:ММ\n\n"+
			"Пример: 15.08.2024 13:00")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func InputVisitTimeEnd(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForTravelVisitTimeEnd
	states.Input.Mx.RUnlock()

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"⏳ Введите дату и время окончания посещения локации в следующем формате:\n\n"+
			"ДД.ММ.ГГГГ ЧЧ:ММ\n\n"+
			"Пример: 15.08.2024 13:00")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent

}
