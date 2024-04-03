package restaurants

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupRestaurants(b *tb.Bot) {
	restaurantsGroup := b.Group()
	restaurantsGroup.Handle(&travel.BtnRestaurants, Menu)
	restaurantsGroup.Handle(&btnBackRestaurants, travel.OpenTravelMenu)
}
