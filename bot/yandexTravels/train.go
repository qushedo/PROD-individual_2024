package yandexTravels

import (
	"backend-qushedo/models"
	nominatim "github.com/doppiogancio/go-nominatim"
	"strings"
	"time"
)

const trainsLink = "https://travel.yandex.ru/trains/"

func GetTrainLink(from models.Location, to models.Location, when time.Time) (string, error) {
	addressFrom, err := nominatim.ReverseGeocode(from.Latitude, from.Longitude, "en")
	if err != nil {
		return "", err
	}
	addressTo, err := nominatim.ReverseGeocode(to.Latitude, to.Longitude, "en")
	if err != nil {
		return "", err
	}

	addressFromCity := strings.ToLower(addressFrom.City)
	addressToCity := strings.ToLower(addressTo.City)
	if addressFromCity == "" {
		addressFromCity = strings.ToLower(strings.Split(addressFrom.County, " ")[0])
	}
	if addressToCity == "" {
		addressToCity = strings.ToLower(strings.Split(addressTo.County, " ")[0])
	}

	dateWhen := when.Format("2006-01-02")
	urlTrains := trainsLink + addressFromCity + "--" + addressToCity + "/?when=" + dateWhen
	return urlTrains, nil
}

func GetTrainLinkByCords(fromLat float64, fromLng float64, toLat float64, toLng float64, when time.Time) (string, error) {
	addressFrom, err := nominatim.ReverseGeocode(fromLat, fromLng, "en")
	if err != nil {
		return "", err
	}
	addressTo, err := nominatim.ReverseGeocode(toLat, toLng, "en")
	if err != nil {
		return "", err
	}

	addressFromCity := strings.ToLower(addressFrom.City)
	addressToCity := strings.ToLower(addressTo.City)
	if addressFromCity == "" {
		addressFromCity = strings.ToLower(strings.Split(addressFrom.County, " ")[0])
	}
	if addressToCity == "" {
		addressToCity = strings.ToLower(strings.Split(addressTo.County, " ")[0])
	}

	dateWhen := when.Format("2006-01-02")
	urlTrains := trainsLink + addressFromCity + "--" + addressToCity + "/?when=" + dateWhen
	return urlTrains, nil
}
