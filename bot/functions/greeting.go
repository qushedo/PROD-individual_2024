package functions

import (
	"backend-qushedo/database"
	"backend-qushedo/keyboards"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	"github.com/doppiogancio/go-nominatim/shared"
	tb "gopkg.in/telebot.v3"
	"log"
)

var (
	selectorStart = &tb.ReplyMarkup{}
	btnStart      = selectorStart.Data("Погнали!", "start")

	menuGeo      = &tb.ReplyMarkup{ResizeKeyboard: true}
	btnManualGeo = menuGeo.Text("🏠 Введу название")
	btnSendGeo   = menuGeo.Location("📍 Отправить геометку")

	selectorGender = &tb.ReplyMarkup{}
	btnMale        = selectorGender.Data("👨 Мужской", "btnMale")
	btnFemale      = selectorGender.Data("👩 Женский", "btnFemale")

	selectorGeoIsCorrect = &tb.ReplyMarkup{}
	btnCorrectGeo        = selectorGeoIsCorrect.Data("✅ Да, всё верно", "geoIsCorrect")
	btnIncorrectGeo      = selectorGeoIsCorrect.Data("❌ Нет, изменить", "geoIsIncorrect")
)

func SetupGreeting(b *tb.Bot) {
	b.Handle("/start", greeting)
	b.Handle(&btnStart, InputName)
	b.Handle(&btnMale, Male)
	b.Handle(&btnFemale, Female)
	b.Handle(&btnManualGeo, InputGeo)
	b.Handle(&btnCorrectGeo, GeoCorrect)
	b.Handle(&btnIncorrectGeo, GeoIncorrect)

}

func greeting(c tb.Context) error {
	var existingUser models.User
	selectorStart.Inline(
		selectorStart.Row(btnStart),
	)

	stateSent := states.Sent.Map[c.Sender().ID]
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, c.Message().ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()

	database.DB.Where("tg_id=?", c.Sender().ID).Find(&existingUser)
	if existingUser.TgId == 0 {
		database.DB.Create(models.User{
			TgId:    c.Sender().ID,
			Name:    "",
			Age:     0,
			Bio:     "",
			Address: "",
		})
	} else if existingUser.TgId != 0 && existingUser.Name != "" && existingUser.Age != 0 && existingUser.Bio != "" && existingUser.Address != "" {
		if c.Message().Payload != "" {
			return UseInviteLink(c)
		}
		return MainMenu(c)
	}
	if c.Message().Payload != "" {
		return c.Reply(
			"Чтобы принять приглашение - пройдите регистрацию\n" +
				"А затем еще раз перейдите по ссылке",
		)
	}

	return c.Send(
		"👋 Привет!\n\n"+
			"Добро пожаловать в Travel Agent 3.0 ✈️\n\n"+
			"Ваш незаменимый помощник для идеального путешествия:\n\n"+
			"👫 Путешествуйте с друзьями\n"+
			"🔍 Найдите попутчиков\n"+
			"📝 Ведите заметки\n"+
			"🗺 Планируйте маршрут\n"+
			"☀️ Узнавайте погоду\n"+
			"🏰 Открывайте новые места\n\n"+
			"И многое другое для вашего приключения!\n\n"+
			"Готовы начать?",
		selectorStart,
	)
}

func InputName(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		log.Println(err)
	}
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForName
	states.Input.Mx.RUnlock()

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(), "Как тебя зовут?")
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func InputAge(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForAge
	states.Input.Mx.RUnlock()

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(), "Сколько тебе лет?")
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func InputBio(c tb.Context) error {
	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForBio
	states.Input.Mx.RUnlock()

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(), "Расскажи немного о себе")
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func ChooseGender(c tb.Context) error {
	selectorGender.Inline(
		selectorGender.Row(btnMale, btnFemale),
	)

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(), "👨👩 Выберите свой пол", selectorGender)
	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func Male(c tb.Context) error {
	_ = c.Delete()
	user, err := database.GetUser(c.Sender().ID)
	if err != nil {
		stateSent := states.Sent.Map[c.Sender().ID]
		sentMsg, errSent := c.Bot().Send(c.Chat(), "Пользователь не найден")
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}
		return InputName(c)
	}
	user.Male = true
	database.DB.Where("tg_id=?", c.Sender().ID).Save(&user)
	return ChooseGeo(c)
}

func Female(c tb.Context) error {
	_ = c.Delete()
	user, err := database.GetUser(c.Sender().ID)
	if err != nil {
		stateSent := states.Sent.Map[c.Sender().ID]
		sentMsg, errSent := c.Bot().Send(c.Chat(), "Пользователь не найден")
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}
		return InputName(c)
	}
	user.Male = false
	database.DB.Where("tg_id=?", c.Sender().ID).Save(&user)
	return ChooseGeo(c)
}

func ChooseGeo(c tb.Context) error {
	menuGeo.Reply(
		menuGeo.Row(btnManualGeo),
		menuGeo.Row(btnSendGeo),
	)

	states.Input.Mx.RLock()
	states.Input.Map[c.Sender().ID] = states.WaitingForGeo
	states.Input.Mx.RUnlock()

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(),
		"🌐 Для начала работы мне нужен доступ к твоей геолокации.\n\n"+
			"Выберите способ поделиться геолокацией:\n\n"+
			"- 📍 Отправить геометку.\n"+
			"- 🏠 Ввести адрес вручную.",
		menuGeo,
	)

	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func InputGeo(c tb.Context) error {

	err := c.Delete()
	if err != nil {
		log.Println(err)
	}

	stateSent := states.Sent.Map[c.Sender().ID]
	msg, err := c.Bot().Send(c.Chat(),
		"🌍 Укажите своё местоположение, чтобы приступить к путешествиям\n\n"+
			"Город, Страна (Например: Москва, Россия) \n\n"+
			"Если не удаётся найти локацию, используйте расширенный формат:\n\n"+
			"Город, Район, Область, Страна", keyboards.EmptyMenu,
	)

	states.Sent.Mx.RLock()
	stateSent.SentMessagesId = append(stateSent.SentMessagesId, msg.ID)
	states.Sent.Map[c.Sender().ID] = stateSent
	states.Sent.Mx.RUnlock()
	return err
}

func IsGeoCorrect(c tb.Context, address shared.Address, coordinate shared.Coordinate) error {
	userId := c.Sender().ID
	selectorGeoIsCorrect.Inline(
		selectorGeoIsCorrect.Row(btnCorrectGeo, btnIncorrectGeo),
	)
	user, err := database.GetUser(userId)
	if err != nil {
		log.Println(err)
	}
	user.Address = address.DisplayName
	user.Latitude = coordinate.Latitude
	user.Longitude = coordinate.Longitude

	database.DB.Where("tg_id=?", userId).Save(&user)

	return c.Send(
		fmt.Sprintf("📍 Ваша геопозиция:\n\n%s\n\n✅ Это верно?", address.DisplayName),
		selectorGeoIsCorrect,
	)

}

func GeoCorrect(c tb.Context) error {
	_, err := database.GetUserHard(c.Sender().ID)
	if err != nil {
		return err
	}

	_ = c.Delete()

	return MainMenu(c)
}

func GeoIncorrect(c tb.Context) error {
	err := c.Delete()
	if err != nil {
		return err
	}

	userId := c.Sender().ID
	user, err := database.GetUser(userId)
	if err != nil {
		log.Println(err)
	}
	user.Address = ""
	database.DB.Where("tg_id=?", userId).Save(&user)
	return ChooseGeo(c)
}
