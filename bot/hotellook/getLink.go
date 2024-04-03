package hotellook

import (
	"net/url"
)

const hotelLookUrl = "https://search.hotellook.com/hotels?=1&"

func GetLink(adults string, checkIn string, checkOut string, children string, destination string) string {
	values := url.Values{}
	values.Add("adults", adults)
	values.Add("checkIn", checkIn)
	values.Add("checkOut", checkOut)
	values.Add("children", children)
	values.Add("currency", "rub")
	values.Add("destination", destination)
	values.Add("language", "ru")

	hotelsLink := hotelLookUrl + values.Encode()
	return hotelsLink
}
