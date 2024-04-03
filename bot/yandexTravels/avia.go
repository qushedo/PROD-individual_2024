package yandexTravels

import (
	"bufio"
	"fmt"
	nominatim "github.com/doppiogancio/go-nominatim"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GetAviaURLOpts struct {
	FromLat  float64
	FromLng  float64
	ToLat    float64
	ToLng    float64
	Adults   string
	Children string
	When     time.Time
}

const ttURL = "https://www.tinkoff.ru/travel/flights/one-way/"

func GetAviaLink(opts GetAviaURLOpts) (string, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat")
	if err != nil {
		return "", err
	}

	addressFrom, err := nominatim.ReverseGeocode(opts.FromLat, opts.FromLng, "en")
	if err != nil {
		return "", err
	}
	addressTo, err := nominatim.ReverseGeocode(opts.ToLat, opts.ToLng, "en")
	if err != nil {
		return "", err
	}

	addressFromCity := addressFrom.City
	addressToCity := addressTo.City
	if addressFromCity == "" {
		addressFromCity = strings.Split(addressFrom.County, " ")[0]
	}
	if addressToCity == "" {
		addressToCity = strings.Split(addressTo.County, " ")[0]
	}

	defer resp.Body.Close()
	var iataFrom, iataTo string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}

		city := strings.Trim(fields[2], `"`)
		iataCode := strings.Trim(fields[4], `"`)
		if city == addressFromCity && iataCode != "\\N" {
			iataFrom = iataCode
		}
		if city == addressToCity && iataCode != "\\N" {
			iataTo = iataCode
		}
		if iataFrom != "" && iataTo != "" {
			break
		}
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("adults", opts.Adults)
	params.Add("children", opts.Children)

	URL := fmt.Sprintf("%s%s-%s/%s?", ttURL, iataFrom, iataTo, opts.When.Format("01-02")) + params.Encode()
	return URL, nil
}
