package routes

import (
	"backend-qushedo/functions"
	"backend-qushedo/functions/POI"
	"backend-qushedo/functions/hotels"
	"backend-qushedo/functions/notes"
	"backend-qushedo/functions/restaurants"
	"backend-qushedo/functions/routeBuilding"
	"backend-qushedo/functions/splitWise"
	"backend-qushedo/functions/tickets"
	"backend-qushedo/functions/travel"
	"backend-qushedo/functions/weather"
	"backend-qushedo/handlers"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	tbMiddleware "gopkg.in/telebot.v3/middleware"
)

func Setup(b *tb.Bot) {
	b.Use(tbMiddleware.AutoRespond())
	b.Use(states.AddToSentState)

	b.Handle("/cancel", functions.Cancel)

	functions.SetupGreeting(b)

	b.Handle(tb.OnText, handlers.OnTextHandler)
	b.Handle(tb.OnLocation, handlers.OnLocation)
	b.Handle(tb.OnCallback, handlers.OnCallback)
	b.Handle(tb.OnMedia, handlers.OnMedia)

	functions.SetupMainMenu(b)
	functions.SetupMyProfile(b)

	travel.SetupTravels(b)
	notes.SetupNotes(b)
	routeBuilding.SetupNotes(b)
	weather.SetupWeather(b)
	POI.SetupPoi(b)
	hotels.SetupHotels(b)
	restaurants.SetupRestaurants(b)
	tickets.SetupTickets(b)
	splitWise.SetupSplitWise(b)

}
