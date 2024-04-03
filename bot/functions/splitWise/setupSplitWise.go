package splitWise

import (
	"backend-qushedo/functions"
	"backend-qushedo/functions/travel"
	tb "gopkg.in/telebot.v3"
)

func SetupSplitWise(b *tb.Bot) {
	splitWiseGroup := b.Group()
	splitWiseGroup.Handle(&travel.BtnSplitWise, Menu)
	splitWiseGroup.Handle(&btnCreateTransaction, CreateTransaction)
	splitWiseGroup.Handle(&btnBackChooseMember, CancelCreating)
	splitWiseGroup.Handle(&btnTransactionsStopAdding, StopCreating)
	splitWiseGroup.Handle(&btnBackTransactions, travel.OpenTravelMenu)
	splitWiseGroup.Handle(&btnBackDebtMenu, functions.Back)
	splitWiseGroup.Handle(&btnBackMustPaidTransactions, Menu)
	splitWiseGroup.Handle(&btnBackMustPaidTransaction, Menu)
	splitWiseGroup.Handle(&btnCloseMustPaidTransaction, CloseDebt)
}
