package travel

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	"github.com/google/uuid"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

var (
	selectorMembersList = &tb.ReplyMarkup{}
	btnGenerateInvite   = selectorMembersList.Data("🔗 Сгенерировать пригласительную ссылку", "btnGenerateInvite")
	btnBackMembersList  = selectorMembersList.Data("< Назад", "btnBackMembersList")

	selectorMemberMenu = &tb.ReplyMarkup{}
	btnMemberKick      = selectorMemberMenu.Data("🚫 Выгнать участника", "btnMemberKick")
	btnBackMemberMenu  = selectorMemberMenu.Data("< Назад", "btnBackMemberMenu")
)

func MembersMenu(c tb.Context) error {
	var (
		travel        models.Travel
		travelOwner   models.User
		travelMembers []models.TravelMember
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
	database.DB.Where("id=?", user.CurrentTravelId).Find(&travel)
	database.DB.Where("tg_id=?", travel.OwnerId).Find(&travelOwner)
	database.DB.Where("travel_id=?", user.CurrentTravelId).Find(&travelMembers)

	btnOwner := selectorMembersList.Data(fmt.Sprintf("%s - Создатель", travelOwner.Name), "travelOwner", fmt.Sprintf("travelOwner_%d", travelOwner.TgId))
	rows = append(rows, selectorMembersList.Row(btnOwner))
	for _, member := range travelMembers {
		btnMember := selectorMembersList.Data(fmt.Sprintf("%s - Участник", member.Name), "travelMember", fmt.Sprintf("travelMember_%d", member.TgId))
		rows = append(rows, selectorMembersList.Row(btnMember))
	}

	rows = append(rows, selectorMembersList.Row(btnGenerateInvite))
	rows = append(rows, selectorMembersList.Row(btnBackMembersList))

	selectorMembersList.Inline(rows...)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"Список участников путешествия", selectorMembersList)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func GenerateInviteLink(c tb.Context) error {
	var travelData models.Travel
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
	database.DB.Where("id=?", user.CurrentTravelId).Find(&travelData)
	if travelData.Id == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Отслеживание не найдено")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}
	inviteData := uuid.New().String()
	newInvite := models.Invite{
		TravelId:      travelData.Id,
		Data:          inviteData,
		TravelOwnerId: user.TgId,
		CreationTime:  time.Now(),
	}
	database.DB.Save(&newInvite)
	inviteLink := fmt.Sprintf("https://t.me/%s?start=%s", c.Bot().Me.Username, inviteData)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		"🔗 *Одноразовая ссылка-приглашение:*"+
			fmt.Sprintf("\n`%s`\n\n", inviteLink)+
			"Ссылка будет активна *24 часа*.", &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return MembersMenu(c)
}

func MemberMenu(c tb.Context, id string) error {
	var (
		member         models.TravelMember
		memberUserData models.User
		genderEmoji    string
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
	memberIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	memberId := int64(memberIdInt)
	user.CurrentMemberId = memberId
	database.DB.Where("tg_id=?", c.Sender().ID).Save(&user)

	database.DB.Where("tg_id=?", memberId).Find(&member)
	if member.TgId == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Участник не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	database.DB.Where("tg_id=?", memberId).Find(&memberUserData)
	if memberUserData.TgId == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Участник не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	selectorMemberMenu.Inline(
		selectorMemberMenu.Row(btnMemberKick),
		selectorMemberMenu.Row(btnBackMemberMenu),
	)

	switch memberUserData.Male {
	case true:
		genderEmoji = "👨"

	case false:
		genderEmoji = "👩"
	}
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("%s %s, %s\n", genderEmoji, memberUserData.Name, functions.DetermineAgeName(memberUserData.Age))+
			fmt.Sprintf("\n📍 Геолокация:\n%s\n", memberUserData.Address)+
			fmt.Sprintf("\n📖 О себе:\n%s\n", memberUserData.Bio)+
			fmt.Sprintf("\n🗓 Присоединился к путешествию:\n%s", member.JoinTime.Format("02.01.2006 15:04")),
		selectorMemberMenu)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func OwnerMenu(c tb.Context, id string) error {
	var (
		ownerUserData models.User
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
	ownerIdInt, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	ownerId := int64(ownerIdInt)
	user.CurrentMemberId = ownerId
	database.DB.Where("tg_id=?", c.Sender().ID).Save(&user)

	database.DB.Where("tg_id=?", ownerId).Find(&user)
	if user.TgId == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Участник не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	database.DB.Where("tg_id=?", ownerId).Find(&ownerUserData)
	if ownerUserData.TgId == 0 {
		sentMsg, errSent := c.Bot().Send(c.Chat(),
			"Участник не найден")

		stateSent := states.Sent.Map[c.Sender().ID]
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}

	selectorMemberMenu.Inline(
		selectorMemberMenu.Row(btnBackMemberMenu),
	)
	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("👑 %s, %s\n", ownerUserData.Name, functions.DetermineAgeName(ownerUserData.Age))+
			fmt.Sprintf("\n📍 Геолокация:\n%s\n", ownerUserData.Address)+
			fmt.Sprintf("\n📖 О себе:\n%s\n", ownerUserData.Bio)+
			"\n👑 Создатель путешествия",
		selectorMemberMenu)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	return errSent
}

func KickFromTravel(c tb.Context) error {
	var member models.TravelMember

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

	database.DB.Where("tg_id=?", user.CurrentMemberId).Find(&member)
	memberName := member.Name
	if member.TgId == 0 {
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

		return MembersMenu(c)
	}
	database.DB.Where("tg_id=?", member.TgId).Delete(&member)

	sentMsg, errSent := c.Bot().Send(c.Chat(),
		fmt.Sprintf("Пользователь %s успешно изгнан из путешествия", memberName))

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	if errSent != nil {
		log.Println(errSent)
	}

	return MembersMenu(c)
}
