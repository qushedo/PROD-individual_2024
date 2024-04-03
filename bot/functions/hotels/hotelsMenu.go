package hotels

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/hotellook"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"
)

var (
	selectorHotels = &tb.ReplyMarkup{}
	btnBackHotels  = selectorHotels.Data("< Назад", "btnBackHotels")
)

func Menu(c tb.Context) error {
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

	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)
	database.DB.Where("tg_id=?", travel.OwnerId).Find(&travelOwner)
	database.DB.Where("travel_id=?", travel.Id).Find(&travelMembers)
	database.DB.Order("visit_time_start ASC").Where("travel_id=?", user.CurrentTravelId).Find(&locations)
	for _, location := range locations {
		if !location.VisitTimeStart.IsZero() && !location.VisitTimeEnd.IsZero() && location.Address != "" {
			countAdults, minors := countAdultsAndMinors(travelMembers)
			if travelOwner.Age >= 18 {
				countAdults++
			} else {
				minors = append(minors, strconv.Itoa(int(travelOwner.Age)))
			}
			minorsString := strings.Join(minors, ",")
			urlHotelLook := hotellook.GetLink(strconv.Itoa(countAdults), location.VisitTimeStart.Format("2006-01-02"), location.VisitTimeEnd.Format("2006-01-02"), minorsString, location.Address)
			webApp := &tb.WebApp{URL: urlHotelLook}
			btnOpenHotelInfo := selectorHotels.WebApp(fmt.Sprintf("%s | %s - %s", location.Address, location.VisitTimeStart.Format("02-01-2006"), location.VisitTimeEnd.Format("02-01-2006")), webApp)
			rows = append(rows, selectorHotels.Row(btnOpenHotelInfo))
		} else {
			database.DB.Where("id=?", location.Id).Delete(&location)
		}
	}

	rows = append(rows, selectorHotels.Row(btnBackHotels))

	selectorHotels.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🏨 Выберите локацию", selectorHotels)

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

func countAdultsAndMinors(members []models.TravelMember) (int, []string) {
	var adults int
	var minorsAges []string
	for _, member := range members {
		if member.Age >= 18 {
			adults++
		} else {
			minorsAges = append(minorsAges, strconv.Itoa(int(member.Age)))
		}
	}
	return adults, minorsAges
}
