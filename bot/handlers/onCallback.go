package handlers

import (
	"backend-qushedo/functions/POI"
	"backend-qushedo/functions/notes"
	"backend-qushedo/functions/splitWise"
	"backend-qushedo/functions/travel"
	"backend-qushedo/functions/weather"
	"backend-qushedo/states"
	tb "gopkg.in/telebot.v3"
	"strings"
)

func OnCallback(c tb.Context) error {
	args := c.Args()

	if len(args) >= 2 {
		data := strings.Split(args[1], "_")
		// It's not my fault, there's a bug in the framework, c.Callback().Unique gives out incorrect information (It does not give it, blyat)

		//I informed the maintainer, they said they would fix it 😁

		unique := data[0]
		id := data[1]

		switch unique {
		case "travel":
			return travel.Menu(c, id)

		case "location":
			return travel.LocationMenu(c, id)

		case "travelMember":
			return travel.MemberMenu(c, id)

		case "travelOwner":
			return travel.OwnerMenu(c, id)

		case "note":
			return notes.NoteMenu(c, id)

		case "privacy":
			return notes.ChangePrivacy(c, id)

		case "weather":
			return weather.GetWeather(c, id)

		case "locationPoi":
			return POI.LocationPoi(c, id)

		case "poi":
			return POI.Info(c, id)

		case "transactionChooseMember":
			return splitWise.TransactionAddMember(c, id)

		case "debt":
			return splitWise.DebtMenu(c, id)

		case "mustPaidTransaction":
			return splitWise.MustPaidTransactionsMenu(c, id)

		case "mustPaidChooseMember":
			return splitWise.MustPaidTransactionMenu(c, id)

		default:
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"По всей видимости, функция еще не готова 😢")

			stateSent := states.Sent.Map[c.Sender().ID]
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			return errSent
		}
	}
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"По всей видимости, функция еще не готова 😢")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}
