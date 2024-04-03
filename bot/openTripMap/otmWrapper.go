package openTripMap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type OtmClient struct {
	ApiKey string
	Cache  map[string]FeatureCollection
}

type Opts struct {
	Lat    float64
	Long   float64
	Rate   string
	Radius string
}

type FeatureCollection struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	ID         string     `json:"id"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
}

type Properties struct {
	Name string  `json:"name"`
	Dist float64 `json:"dist"`
}

func (otmClient *OtmClient) GetPlaces(opts Opts) (FeatureCollection, error) {
	url := "https://api.opentripmap.com/0.1/ru/places/radius"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}

	stringLong := fmt.Sprintf("%f", opts.Long)
	stringLat := fmt.Sprintf("%f", opts.Lat)

	query := req.URL.Query()
	query.Set("apikey", otmClient.ApiKey)
	query.Set("lang", "ru")
	query.Set("radius", opts.Radius)
	query.Set("rate", opts.Rate)
	query.Set("lon", stringLong)
	query.Set("lat", stringLat)
	query.Set("kinds", "interesting_places")

	req.URL.RawQuery = query.Encode()
	if result, ok := otmClient.Cache[req.URL.RawQuery]; ok {
		return result, nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return FeatureCollection{}, err
	}
	defer resp.Body.Close()

	var fc FeatureCollection
	err = json.NewDecoder(resp.Body).Decode(&fc)
	if err != nil {
		return FeatureCollection{}, err
	}

	sort.Slice(fc.Features, func(i, j int) bool {
		return fc.Features[i].Properties.Dist < fc.Features[j].Properties.Dist
	})

	otmClient.Cache[req.URL.RawQuery] = fc

	return fc, nil
}

func (otmClient *OtmClient) GetFeatureByID(id string, opts Opts) (*Feature, bool) {
	url := "https://api.opentripmap.com/0.1/ru/places/radius"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}

	stringLong := fmt.Sprintf("%f", opts.Long)
	stringLat := fmt.Sprintf("%f", opts.Lat)

	query := req.URL.Query()
	query.Set("apikey", otmClient.ApiKey)
	query.Set("lang", "ru")
	query.Set("radius", opts.Radius)
	query.Set("rate", opts.Rate)
	query.Set("lon", stringLong)
	query.Set("lat", stringLat)
	query.Set("kinds", "interesting_places")

	req.URL.RawQuery = query.Encode()

	if fc, ok := otmClient.Cache[req.URL.RawQuery]; ok {
		for _, feature := range fc.Features {
			if feature.ID == id {
				return &feature, true
			}
		}
	}

	return nil, false
}
