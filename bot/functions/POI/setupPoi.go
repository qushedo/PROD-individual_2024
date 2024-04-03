package POI

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupPoi(b *tb.Bot) {
	poiGroup := b.Group()
	poiGroup.Handle(&travel.BtnPoi, Menu)
	poiGroup.Handle(&btnBackLocationsPoi, travel.OpenTravelMenu)
	poiGroup.Handle(&btnBackPoi, Menu)
	b.Handle(&btnBackPoiInfo, BackPoiInfo)
}
