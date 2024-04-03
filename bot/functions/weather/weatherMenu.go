package weather

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/functions/travel"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"encoding/json"
	"fmt"
	"github.com/innotechdevops/openmeteo"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

var (
	selectorWeatherLocations = &tb.ReplyMarkup{}
	btnBackWeather           = selectorWeatherLocations.Data("< Назад", "btnBackWeather")
)

type weatherData struct {
	Daily struct {
		Time                     []string  `json:"time"`
		Temperature2mMax         []float64 `json:"temperature_2m_max"`
		Temperature2mMin         []float64 `json:"temperature_2m_min"`
		ApparentTemperatureMax   []float64 `json:"apparent_temperature_max"`
		ApparentTemperatureMin   []float64 `json:"apparent_temperature_min"`
		PrecipitationProbability []int     `json:"precipitation_probability_mean"`
	} `json:"daily"`
}

func Menu(c tb.Context) error {
	var (
		locations []models.Location
		rows      []tb.Row
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

	database.DB.Order("visit_time_start ASC").Where("travel_id=?", user.CurrentTravelId).Find(&locations)
	for _, location := range locations {
		if !location.VisitTimeStart.IsZero() && !location.VisitTimeEnd.IsZero() && location.Address != "" {
			btnWeatherLocation := selectorWeatherLocations.Data(fmt.Sprintf("%s - %s", location.Address, location.VisitTimeStart.Format("02.01.2006")), "weatherLocation", fmt.Sprintf("weather_%d", location.Id))
			rows = append(rows, selectorWeatherLocations.Row(btnWeatherLocation))
		} else {
			database.DB.Where("id=?", location.Id).Delete(&location)
		}
	}

	rows = append(rows, selectorWeatherLocations.Row(btnBackWeather))

	selectorWeatherLocations.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите локацию для получения прогноза погоды", selectorWeatherLocations)

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

func GetWeather(c tb.Context, id string) error {
	var (
		data     weatherData
		location models.Location
		foreCast string
	)

	locationIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	locationId := uint(locationIdInt)
	database.DB.Where("id=?", locationId).Find(&location)

	param := openmeteo.Parameter{
		Latitude:     openmeteo.Float32(float32(location.Latitude)),
		Longitude:    openmeteo.Float32(float32(location.Longitude)),
		ForecastDays: openmeteo.Int(16),
		Daily: &[]string{
			openmeteo.DailyTemperature2mMax,
			openmeteo.DailyTemperature2mMin,
			openmeteo.DailyApparentTemperatureMax,
			openmeteo.DailyApparentTemperatureMin,
			openmeteo.DailyPrecipitationProbabilityMean,
		},
	}

	m := openmeteo.New()
	resp, err := m.Execute(param)

	errUnmarshal := json.Unmarshal([]byte(resp), &data)
	if errUnmarshal != nil {
		fmt.Println("Error decoding JSON:", err)
	}

	foreCast = fmt.Sprintf("Прогноз погоды в %s\n", location.Address)

	startTime := location.VisitTimeStart
	endTime := location.VisitTimeEnd

	foundWeather := false
	for i, t := range data.Daily.Time {
		timeInData, _ := time.Parse("2006-01-02", t)
		if timeInData.After(startTime.AddDate(0, 0, -1)) && timeInData.Before(endTime.AddDate(0, 0, 1)) {
			foundWeather = true

			avgTemp := (data.Daily.Temperature2mMax[i] + data.Daily.Temperature2mMin[i]) / 2
			avgApparentTemp := (data.Daily.ApparentTemperatureMax[i] + data.Daily.ApparentTemperatureMin[i]) / 2

			dailyForecast :=
				fmt.Sprintf("\n🗓 Дата: %s\n", t) +
					fmt.Sprintf("- Средняя температура: %.2f°C\n", avgTemp) +
					fmt.Sprintf("- Средняя ощущаемая температура: %.2f°C\n", avgApparentTemp) +
					fmt.Sprintf("- Вероятность осадков: %d%%\n", data.Daily.PrecipitationProbability[i])

			foreCast += dailyForecast
		}
	}

	if !foundWeather {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Не найден прогноз погоды")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}
		return Menu(c)
	}

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		foreCast)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}
	return Menu(c)
}

func BackWeather(c tb.Context) error {
	stateSent := states.Sent.Map[c.Sender().ID]
	stateSent.Delete(c)
	return travel.OpenTravelMenu(c)
}
