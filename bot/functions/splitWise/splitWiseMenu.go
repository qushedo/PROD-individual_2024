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
)

var (
	selectorSplitWise    = &tb.ReplyMarkup{}
	btnCreateTransaction = selectorSplitWise.Data("💳 Новая транзакция", "bntCreateTransaction")
	btnBackTransactions  = selectorSplitWise.Data("< Назад", "btnBackTransactions")

	selectorDebtMenu = &tb.ReplyMarkup{}
	btnBackDebtMenu  = selectorDebtMenu.Data("< Назад", "btnBackDebtMenu")

	selectorMustPaidTransactions = &tb.ReplyMarkup{}
	btnBackMustPaidTransactions  = selectorMustPaidTransactions.Data("< Назад", "btnBackMustPaidTransactions")

	selectorMustPaidTransaction = &tb.ReplyMarkup{}
	btnCloseMustPaidTransaction = selectorMustPaidTransaction.Data("💰 Закрыть долг", "btnCloseMustPaidTransaction")
	btnBackMustPaidTransaction  = selectorMustPaidTransaction.Data("< Назад", "btnBackMustPaidTransaction")
)

func Menu(c tb.Context) error {
	var (
		transactions         []models.Transaction
		mustPaidTransactions []models.Transaction
		debts                []models.Transaction
		mustPaidSum          int
		debtsSum             int
		rows                 []tb.Row
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

	user.CurrentTransactionUserId = 0
	database.DB.Where("tg_id =?", user.TgId).Save(&user)

	database.DB.Where("owner_id = ? AND travel_id = ?", user.TgId, user.CurrentTravelId).Find(&mustPaidTransactions)
	for _, mustPaidTransaction := range mustPaidTransactions {
		if len(mustPaidTransaction.Participants) > 0 {
			transactionSum := sumDebts(mustPaidTransaction)
			mustPaidSum += transactionSum
			btnMustPaidTransaction := selectorSplitWise.Data(fmt.Sprintf("Транзакция %s. Вам должны %dр.", mustPaidTransaction.CreatedAt.Format("02.01.2006"), transactionSum), "mustPaidTransaction", fmt.Sprintf("mustPaidTransaction_%d", mustPaidTransaction.Id))
			rows = append(rows, selectorSplitWise.Row(btnMustPaidTransaction))
		} else {
			database.DB.Where("id = ?", mustPaidTransaction.Id).Delete(&mustPaidTransaction)
		}
	}

	database.DB.Where("travel_id = ?", user.CurrentTravelId).Find(&transactions)
	for _, transaction := range transactions {
		for _, debt := range transaction.Participants {
			if debt.ParticipantId == user.TgId {
				debts = append(debts, transaction)
			}
		}
	}

	for _, debt := range debts {
		userDebtSum := userDebtInTransaction(user.TgId, debt)
		debtsSum += userDebtSum
		btnDebt := selectorSplitWise.Data(fmt.Sprintf("Транзакция %s. Вы должны %dр.", debt.CreatedAt.Format("02.01.2006"), userDebtSum), "debt", fmt.Sprintf("debt_%d", debt.Id))
		rows = append(rows, selectorSplitWise.Row(btnDebt))
	}

	splitWiseDesc :=
		fmt.Sprintf("🤑 Вам должны %dр.\n", mustPaidSum) +
			fmt.Sprintf("💸 Вы должны %dр.", debtsSum)

	rows = append(rows, selectorSplitWise.Row(btnCreateTransaction))
	rows = append(rows, selectorSplitWise.Row(btnBackTransactions))
	selectorSplitWise.Inline(rows...)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		splitWiseDesc, selectorSplitWise)

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

func MustPaidTransactionsMenu(c tb.Context, id string) error {
	var (
		transaction    models.Transaction
		debtUsers      []models.User
		rows           []tb.Row
		participantIds []int64
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

	_ = c.Delete()
	database.DB.Where("id = ?", id).Find(&transaction)
	for _, debt := range transaction.Participants {
		participantIds = append(participantIds, debt.ParticipantId)
	}
	database.DB.Where("tg_id IN (?)", participantIds).Find(&debtUsers)

	for _, debtUser := range debtUsers {
		amount := debtAmount(transaction, debtUser.TgId)
		btnMustPaidTransaction := selectorMustPaidTransactions.Data(fmt.Sprintf("%s - %dр.", debtUser.Name, amount), "mustPaidChooseMember", fmt.Sprintf("mustPaidChooseMember_%d", debtUser.TgId))
		rows = append(rows, selectorMustPaidTransactions.Row(btnMustPaidTransaction))
	}

	rows = append(rows, selectorMustPaidTransactions.Row(btnBackMustPaidTransactions))
	selectorMustPaidTransactions.Inline(rows...)

	user.CurrentTransactionId = transaction.Id
	database.DB.Where("tg_id =?", user.TgId).Save(&user)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Выберите пользователя", selectorMustPaidTransactions)

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

func MustPaidTransactionMenu(c tb.Context, id string) error {
	var (
		transaction models.Transaction
		debtUser    models.User
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

	_ = c.Delete()
	database.DB.Where("id = ?", user.CurrentTransactionId).Find(&transaction)

	intUserId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	userId := int64(intUserId)

	amount := debtAmount(transaction, userId)
	database.DB.Where("tg_id = ?", userId).Find(&debtUser)
	selectorMustPaidTransaction.Inline(selectorMustPaidTransaction.Row(selectorMustPaidTransaction.Data("💰 Закрыть долг", "btnCloseMustPaidTransaction", fmt.Sprintf("%d", userId))),
		selectorMustPaidTransaction.Row(btnBackMustPaidTransaction),
	)
	transactionDesc :=
		fmt.Sprintf("%s должен вам %dр.\n", debtUser.Name, amount) +
			fmt.Sprintf("Дата: %s\n", transaction.CreatedAt.Format("02.01.2006")) +
			fmt.Sprintf("Время: %s", transaction.CreatedAt.Format("15:04"))

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		transactionDesc, selectorMustPaidTransaction)

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

func CloseDebt(c tb.Context) error {
	var (
		transaction models.Transaction
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

	intUserId, err := strconv.Atoi(c.Callback().Data)
	if err != nil {
		return err
	}
	userId := int64(intUserId)

	database.DB.Where("id =?", user.CurrentTransactionId).Find(&transaction)
	removeDebtByParticipantId(&transaction, userId)

	database.DB.Where("id = ?", transaction.Id).Save(&transaction)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Долг успешно закрыт")

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return Menu(c)
}

func DebtMenu(c tb.Context, id string) error {
	var (
		transaction models.Transaction
		owner       models.User
	)
	database.DB.Where("id = ?", id).Find(&transaction)
	database.DB.Where("tg_id =?", transaction.OwnerId).Find(&owner)

	amount := debtAmount(transaction, c.Sender().ID)
	debtDesc :=
		fmt.Sprintf("💸 Вы должны пользователю %s %dр.\n", owner.Name, amount) +
			fmt.Sprintf("Дата: %s\n", transaction.CreatedAt.Format("02.01.2006")) +
			fmt.Sprintf("Время: %s", transaction.CreatedAt.Format("15:04"))

	selectorDebtMenu.Inline(
		selectorDebtMenu.Row(btnBackDebtMenu),
	)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		debtDesc, selectorDebtMenu)

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

func debtAmount(transaction models.Transaction, userId int64) int {
	for _, debt := range transaction.Participants {
		if debt.ParticipantId == userId {
			return debt.Amount
		}

	}
	return 0
}

func sumDebts(transaction models.Transaction) int {
	var sum int
	for _, debt := range transaction.Participants {
		sum += debt.Amount
	}
	return sum
}

func userDebtInTransaction(userId int64, transaction models.Transaction) int {
	var userDebt int
	for _, debt := range transaction.Participants {
		if debt.ParticipantId == userId {
			userDebt += debt.Amount
		}
	}
	return userDebt
}

func removeDebtByParticipantId(transaction *models.Transaction, participantId int64) {
	var updatedParticipants []models.Debt
	for _, debt := range transaction.Participants {
		if debt.ParticipantId != participantId {
			updatedParticipants = append(updatedParticipants, debt)
		}
	}
	transaction.Participants = updatedParticipants
}
