package hotels

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupHotels(b *tb.Bot) {
	hotelsGroup := b.Group()
	hotelsGroup.Handle(&travel.BtnHotels, Menu)
	hotelsGroup.Handle(&btnBackHotels, travel.OpenTravelMenu)
}
