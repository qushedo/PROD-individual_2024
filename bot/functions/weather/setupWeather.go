package weather

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupWeather(b *tb.Bot) {
	weatherGroup := b.Group()
	weatherGroup.Handle(&travel.BtnWeather, Menu)
	weatherGroup.Handle(&btnBackWeather, BackWeather)
}
