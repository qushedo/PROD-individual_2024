package tickets

import (
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupTickets(b *tb.Bot) {
	ticketsGroup := b.Group()
	ticketsGroup.Handle(&travel.BtnTickets, ChooseTransport)
	ticketsGroup.Handle(&btnAvia, AviaMenu)
	ticketsGroup.Handle(&btnTrain, TrainMenu)
	ticketsGroup.Handle(&btnBus, BusMenu)
	ticketsGroup.Handle(&btnBackTicketsTransport, travel.OpenTravelMenu)
	ticketsGroup.Handle(&btnBackTicketsAvia, ChooseTransport)
	ticketsGroup.Handle(&btnBackTicketsTrain, ChooseTransport)
	ticketsGroup.Handle(&btnBackTicketsBus, ChooseTransport)
}
