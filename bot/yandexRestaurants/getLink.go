package yandexRestaurants

import "fmt"

func GetLink(city string) string {
	return fmt.Sprintf("https://yandex.ru/maps/213/moscow/search/%s,рестораны/", city)
}
