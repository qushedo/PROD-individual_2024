package tests

import (
	"backend-qushedo/graphhopper"
	"backend-qushedo/openTripMap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOtmWrapper(t *testing.T) {
	openTripMap.NewClient("5ae2e3f221c38a28845f05b6d98163ab951f81b33208c8ed5b0bd7a0")
	opts := openTripMap.Opts{
		Lat:    55.625578,
		Long:   37.6063916,
		Rate:   "3, 3h, 2, 2h",
		Radius: "10000",
	}
	fc, err := openTripMap.Otm.GetPlaces(opts)

	wantName := "Храм Покрова Пресвятой Богородицы на Городне"
	wantDist := 2831.34290671

	assert.NoError(t, err)

	assert.Equal(t, fc.Features[0].Properties.Name, wantName)
	assert.Equal(t, fc.Features[0].Properties.Dist, wantDist)

	feature, ok := openTripMap.Otm.GetFeatureByID("11556532", opts)
	assert.True(t, ok)

	wantName = "Царицыно"
	wantDist = 3982.10395505

	assert.Equal(t, feature.Properties.Name, wantName)
	assert.Equal(t, feature.Properties.Dist, wantDist)
}

func TestGraphWrapper(t *testing.T) {
	var ghLocs []graphhopper.GhLocation

	ghLocs = append(ghLocs, graphhopper.GhLocation{
		Address:   "Озёрная улица, Ковров, городской округ Ковров, Владимирская область, Центральный федеральный округ, 601901, Россия",
		Latitude:  56.355386100000004,
		Longitude: 37.6174782,
	})

	ghLocs = append(ghLocs, graphhopper.GhLocation{
		Address:   "Москва, Россия",
		Latitude:  55.7505412,
		Longitude: 37.6174782,
	})

	_, err := graphhopper.GetLink(ghLocs)
	assert.NoError(t, err)
}

// God, 5:30 a.m. on March 26th, what are you wanting from me?
