package functions

import (
	"backend-qushedo/database"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"time"
)

func UseInviteLink(c tb.Context) error {
	var (
		invite     models.Invite
		member     models.TravelMember
		travelData models.Travel
	)
	inviteData := c.Message().Payload

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

		return InputName(c)
	}
	database.DB.Where("data=?", inviteData).Find(&invite)
	if invite.Id == 0 {
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Ссылка-приглашение недействительна")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return MainMenu(c)
	}

	if user.TgId == invite.TravelOwnerId {
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Вы не можете присоединиться к путешествию, создателем которого являетесь")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return MainMenu(c)
	}

	if invite.CreationTime.Sub(time.Now()).Hours() > 24 {
		database.DB.Delete(&invite)

		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Ссылка-приглашение недействительна")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return MainMenu(c)

	}
	database.DB.Where("id=?", invite.TravelId).Find(&travelData)
	if travelData.Id == 0 {
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Путешествие не найдено")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return MainMenu(c)
	}

	database.DB.Where("tg_id=? AND travel_id=?", user.TgId, invite.TravelId).Find(&member)
	if member.TgId != 0 {
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Вы уже являетесь участником этого путешествия")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}

		return MainMenu(c)
	}

	newMember := models.TravelMember{
		Name:     user.Name,
		Age:      user.Age,
		TgId:     user.TgId,
		TravelId: invite.TravelId,
		JoinTime: time.Now(),
	}
	database.DB.Create(&newMember)

	sendNotification(c, invite.TravelId, user.Name)
	database.DB.Delete(&invite)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("Вы успешно стали участником путешествия %s", travelData.Name))

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return MainMenu(c)
} // I hate import cycles

func sendNotification(c tb.Context, travelId uint, memberName string) {
	var (
		travel        models.Travel
		travelOwner   models.User
		travelMembers []models.TravelMember
	)
	database.DB.Where("id=?", travelId).Find(&travel)
	database.DB.Where("tg_id=?", travel.OwnerId).Find(&travelOwner)
	database.DB.Where("travel_id=?", travel.Id).Find(&travelMembers)

	ownerChat, _ := c.Bot().ChatByID(travelOwner.TgId)
	sentMsg, _ := c.Bot().Send(ownerChat,
		fmt.Sprintf("✨ %s присоединился к путешествию", memberName))

	stateSent := states.Sent.Map[travelOwner.TgId]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[travelOwner.TgId] = stateSent
	states.Sent.Mx.RUnlock()

	for _, member := range travelMembers {
		memberChat, _ := c.Bot().ChatByID(member.TgId)
		sentMsg, _ = c.Bot().Send(memberChat,
			fmt.Sprintf("✨ %s присоединился к путешествию", memberName))

		stateSent = states.Sent.Map[travelOwner.TgId]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[travelOwner.TgId] = stateSent
		states.Sent.Mx.RUnlock()
	}
}
