package graphhopper

import (
	"backend-qushedo/models"
	"errors"
	"fmt"
	"net/url"
)

const (
	graphhopperUrl = "https://graphhopper.com/maps/?"
)

type GhLocation struct {
	Address   string
	Latitude  float64
	Longitude float64
}

func GetLinkByLocations(locations []models.Location) (string, error) {
	values := url.Values{}
	if len(locations) >= 2 {
		for _, location := range locations {
			values.Add("point", fmt.Sprintf("%f,%f_%s", location.Latitude, location.Longitude, location.Address))
		}
	} else {
		return "", errors.New("Для построения маршрута в путешествии должно быть как минимум 2 локации")
	}

	values.Add("profile", "car")
	values.Add("layer", "Omniscale")

	mapUrl := graphhopperUrl + values.Encode()
	return mapUrl, nil
}

func GetLink(locations []GhLocation) (string, error) {
	values := url.Values{}
	if len(locations) >= 2 {
		for _, location := range locations {
			values.Add("point", fmt.Sprintf("%f,%f_%s", location.Latitude, location.Longitude, location.Address))
		}
	} else {
		return "", errors.New("Для построения маршрута в путешествии должна быть как минимум 1 локация")
	}

	values.Add("profile", "car")
	values.Add("layer", "Omniscale")

	mapUrl := graphhopperUrl + values.Encode()
	return mapUrl, nil
}
