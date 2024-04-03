package database

import (
	"backend-qushedo/models"
	"errors"
)

func GetUser(tgId int64) (models.User, error) {
	var user models.User
	DB.Where("tg_id=?", tgId).Find(&user)
	if user.TgId == 0 {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}

func GetUserHard(tgId int64) (models.User, error) {
	var user models.User
	DB.Where("tg_id=?", tgId).Find(&user)
	if user.TgId == 0 || user.Name == "" || user.Age == 0 || user.Bio == "" {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}

func GetTravelLocations(travelId uint) []models.Location {
	var locations []models.Location
	DB.Order("visit_time_start ASC").Where("travel_id=?", travelId).Find(&locations)
	return locations
}

func GetTravelFirstLocation(travelId uint) models.Location {
	var location models.Location
	DB.Order("visit_time_start ASC").Where("travel_id=?", travelId).First(&location)
	return location
}
