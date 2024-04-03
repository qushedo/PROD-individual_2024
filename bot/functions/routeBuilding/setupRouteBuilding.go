package routeBuilding

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupNotes(b *tb.Bot) {
	routeBuildingGroup := b.Group()

	routeBuildingGroup.Handle(&travel.BtnBuildRoute, ChooseMode)
	routeBuildingGroup.Handle(&btnRouteTravel, BuildTravelRoute)
	routeBuildingGroup.Handle(&btnRouteToStartPos, BuildRouteToStartPos)
	routeBuildingGroup.Handle(&btnBackRoute, travel.OpenTravelMenu)
	routeBuildingGroup.Handle(&btnBackRouteChoose, travel.OpenTravelMenu)
}
