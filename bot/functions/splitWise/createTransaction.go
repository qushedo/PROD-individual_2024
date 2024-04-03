package splitWise

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

var (
	selectorChooseMember      = &tb.ReplyMarkup{}
	btnBackChooseMember       = selectorChooseMember.Data("< Отменить", "btnBackChooseMember")
	btnTransactionsStopAdding = selectorChooseMember.Data("🛑 Хватит", "btnTransactionsStopAdding")
)

func CreateTransaction(c tb.Context) error {
	_ = c.Delete()

	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}
	newTransaction := models.Transaction{
		OwnerId:   user.TgId,
		TravelId:  user.CurrentTravelId,
		CreatedAt: time.Now(),
	}
	database.DB.Create(&newTransaction)
	user.CurrentTransactionId = newTransaction.Id
	database.DB.Where("tg_id = ?", user.TgId).Save(&user)
	return ChooseMember(c, []int64{})
}

func ChooseMember(c tb.Context, withoutUsers []int64) error {
	var (
		owner         models.User
		travel        models.Travel
		travelMembers []models.TravelMember
		transaction   models.Transaction
		rows          []tb.Row
	)
	_ = c.Delete()
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	database.DB.Where("id = ?", user.CurrentTravelId).Find(&travel)
	database.DB.Where("travel_id = ?", travel.Id).Find(&travelMembers)
	database.DB.Where("tg_id = ?", travel.OwnerId).Find(&owner)
	database.DB.Where("id = ?", user.CurrentTransactionId).Find(&transaction)

	if user.TgId != owner.TgId && !checkUserInWithout(owner, withoutUsers) {
		btnOwner := selectorChooseMember.Data(fmt.Sprintf("%s", owner.Name), "transactionsChooseMember", fmt.Sprintf("transactionChooseMember_%d", owner.TgId))
		rows = append(rows, selectorChooseMember.Row(btnOwner))
	}

	for _, member := range travelMembers {
		if user.TgId != member.TgId && !checkMemberInWithout(member, withoutUsers) {
			btnMember := selectorChooseMember.Data(fmt.Sprintf("%s", member.Name), "transactionsChooseMember", fmt.Sprintf("transactionChooseMember_%d", member.TgId))
			rows = append(rows, selectorChooseMember.Row(btnMember))
		}
	}

	if len(transaction.Participants) > 0 {
		rows = append(rows, selectorChooseMember.Row(btnTransactionsStopAdding))
	}
	rows = append(rows, selectorChooseMember.Row(btnBackChooseMember))

	selectorChooseMember.Inline(rows...)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите пользователя который вам должен", selectorChooseMember)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return errSent
}

func TransactionAddMember(c tb.Context, id string) error {

	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}

	userIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	userId := int64(userIdInt)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Введите сумму задолженности")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}
	user.CurrentTransactionUserId = userId
	database.DB.Where("tg_id = ?", user.TgId).Save(&user)

	states.Input.Mx.RLock()
	states.Input.Map[user.TgId] = states.WaitingForTransactionAmount
	states.Input.Mx.RUnlock()

	return errSent
}

func CancelCreating(c tb.Context) error {
	var (
		transaction models.Transaction
	)
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}
	database.DB.Where("id = ?", user.CurrentTransactionId).Delete(&transaction)
	user.CurrentTransactionId = 0
	user.CurrentTransactionUserId = 0
	database.DB.Where("tg_id =?", user.TgId).Save(&user)

	states.Input.Mx.RLock()
	states.Input.Map[user.TgId] = 0
	states.Input.Mx.RUnlock()

	stateSent := states.Sent.Map[c.Sender().ID]
	stateSent.Delete(c)

	return Menu(c)
}

func StopCreating(c tb.Context) error {
	var (
		transaction models.Transaction
	)
	user, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Пользователь не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return functions.InputName(c)
	}
	database.DB.Where("id = ?", user.CurrentTransactionId).Find(&transaction)
	notifyTransactionParticipants(c, transaction)
	user.CurrentTransactionId = 0
	user.CurrentTransactionUserId = 0
	database.DB.Where("tg_id = ?", user.TgId).Save(&user)

	return Menu(c)
}

func notifyTransactionParticipants(c tb.Context, transaction models.Transaction) {
	for _, member := range transaction.Participants {
		chat, _ := c.Bot().ChatByID(member.ParticipantId)
		sentMsg, errSent := c.Bot().Send(chat,
			"💸 Вы стали участником новой транзакции\n"+
				fmt.Sprintf("Ваш долг составляет %d.", debtAmount(transaction, member.ParticipantId)))

		stateSent := states.Sent.Map[member.ParticipantId]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[member.ParticipantId] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}
	}
}

func checkMemberInWithout(member models.TravelMember, without []int64) bool {
	exclude := false
	for _, withoutUser := range without {
		if member.TgId == withoutUser {
			exclude = true
			break
		}
	}
	return exclude
}

func checkUserInWithout(member models.User, withoutUsers []int64) bool {
	exclude := false
	for _, withoutUser := range withoutUsers {
		if member.TgId == withoutUser {
			exclude = true
			break
		}
	}
	return exclude
}
