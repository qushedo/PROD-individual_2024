package handlers

import (
	"backend-qushedo/database"
	"backend-qushedo/functions"
	"backend-qushedo/functions/notes"
	"backend-qushedo/functions/splitWise"
	"backend-qushedo/functions/travel"
	"backend-qushedo/models"
	"backend-qushedo/states"
	"fmt"
	"github.com/bregydoc/gtranslate"
	nominatim "github.com/doppiogancio/go-nominatim"
	"golang.org/x/text/language"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// OnTextHandler Сомнительно..., но окээй
func OnTextHandler(c tb.Context) error {

	userId := c.Sender().ID
	state := states.Input.Map[userId]
	stateSent := states.Sent.Map[c.Sender().ID]
	msg := c.Message().Text
	user, err := database.GetUser(userId)
	if err != nil {
		sentMsg, errSent := c.Bot().Send(c.Chat(), "Пользователь не найден")
		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		if errSent != nil {
			log.Println(errSent)
		}
		return functions.InputName(c)
	}

	switch state {
	case states.WaitingForName:
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		if utf8.RuneCountInString(msg) >= 2 && utf8.RuneCountInString(msg) <= 15 && checkLetters(msg) {
			user.Name = strings.TrimSpace(msg)
			database.DB.Where("tg_id=?", userId).Save(&user)

			return functions.InputAge(c)

		} else {

			sentMsg, errSent := c.Bot().Reply(c.Message(), "Что то с твоим именем не так, попробуй еще раз")
			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.InputName(c)
		}

	case states.WaitingForAge:
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		intAge, err := strconv.Atoi(msg)
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Возраст должен быть числом\n"+
					"Попробуй еще раз")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.InputAge(c)
		}
		if utf8.RuneCountInString(msg) < 4 && utf8.RuneCountInString(msg) > 0 && intAge >= 10 && intAge <= 118 {
			user.Age = uint(intAge)
			database.DB.Where("tg_id=?", userId).Save(&user)

			return functions.InputBio(c)

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Ты либо слишком молод либо слишком стар для путешествий\n"+
					"Попробуй еще раз")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.InputAge(c)
		}
	case states.WaitingForBio:
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		if utf8.RuneCountInString(msg) >= 4 && utf8.RuneCountInString(msg) <= 200 {
			user.Bio = strings.TrimSpace(msg)
			database.DB.Where("tg_id=?", userId).Save(&user)

			return functions.ChooseGender(c)

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				`Длина "О себе" должна быть от 4 до 200 символов`+"\n"+
					`Попробуй еще раз`)

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.InputBio(c)
		}

	case states.WaitingForGeo:

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		translated, err := gtranslate.Translate(msg, language.Russian, language.English)
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Произошла ошибка при поиске локации")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.InputGeo(c)
		}
		coords, err := nominatim.Geocode(strings.ReplaceAll(translated, " ", ""))
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Локация не найдена")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return functions.ChooseGeo(c)
		} else {
			address, err := nominatim.ReverseGeocode(coords.Latitude, coords.Longitude, "ru")
			if err != nil {
				log.Println(err)
			}

			return functions.IsGeoCorrect(c, *address, *coords)
		}

	case states.WaitingForTravelName:
		var travelData models.Travel

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		if utf8.RuneCountInString(msg) >= 3 && utf8.RuneCountInString(msg) <= 30 {
			database.DB.Where("owner_id=? AND name=?", userId, msg).First(&travelData) // I do not know how to make it so that it does not output an error to the console, fuck it
			if travelData.Id == 0 {
				database.DB.Where("owner_id=? AND name=?", userId, "").First(&travelData)
				travelData.Name = msg
				database.DB.Where("id=?", travelData.Id).Save(&travelData)
				return travel.InputTravelDesc(c)

			} else {
				sentMsg, errSent := c.Bot().Send(c.Chat(),
					"Название путешествия должно быть уникальным")

				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return travel.InputTravelName(c)
			}

		} else {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Название путешествия должно быть от 3 до 30 символов")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputTravelName(c)
		}

	case states.WaitingForTravelDescription:
		var travelData models.Travel

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		if utf8.RuneCountInString(msg) >= 4 && utf8.RuneCountInString(msg) <= 200 {
			database.DB.Where("owner_id=? AND description=?", userId, "").First(&travelData)
			travelData.Description = msg
			database.DB.Where("id=?", travelData.Id).Save(&travelData)

			sentMsg, errSent := c.Bot().Send(c.Chat(),
				fmt.Sprintf(`Путешествие "%s" успешно создано`, travelData.Name))

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.MyTravels(c)
		} else {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Описание путешествия должно быть от 4 до 200 символов")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputTravelDesc(c)
		}

	case states.WaitingForTravelNameEdit:
		var travelData models.Travel

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		if utf8.RuneCountInString(msg) >= 3 && utf8.RuneCountInString(msg) <= 20 {
			database.DB.Where("owner_id=? AND name=?", userId, msg).First(&travelData) // I do not know how to make it so that it does not output an error to the console, fuck it
			if travelData.Id == 0 {
				database.DB.Where("owner_id=? AND id=?", userId, user.CurrentTravelId).First(&travelData)
				travelData.Name = msg
				database.DB.Where("id=?", travelData.Id).Save(&travelData)

				travelIdStr := strconv.Itoa(int(travelData.Id))
				if err != nil {
					log.Println(err)
				}

				return travel.Menu(c, travelIdStr)

			} else {
				sentMsg, errSent := c.Bot().Send(c.Chat(),
					"Название путешествия должно быть уникальным")

				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return travel.EditTravelName(c)
			}

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Название путешествия должно быть от 3 до 20 символов")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.EditTravelName(c)
		}

	case states.WaitingForTravelDescEdit:
		var travelData models.Travel

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		if utf8.RuneCountInString(msg) >= 4 && utf8.RuneCountInString(msg) <= 200 {
			database.DB.Where("owner_id=? AND id=?", userId, user.CurrentTravelId).First(&travelData)
			travelData.Description = msg
			database.DB.Where("id=?", travelData.Id).Save(&travelData)

			travelIdStr := strconv.Itoa(int(travelData.Id))
			if err != nil {
				log.Println(err)
			}

			return travel.Menu(c, travelIdStr)

		} else {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Описание путешествия должно быть от 4 до 200 символов")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.EditTravelDesc(c)
		}

	case states.WaitingForTravelLocation:
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		translated, err := gtranslate.Translate(msg, language.Russian, language.English)
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Произошла ошибка при поиске локации")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputTravelLocation(c)
		}
		coords, err := nominatim.Geocode(strings.ReplaceAll(translated, " ", ""))
		if err != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Произошла ошибка при поиске локации")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputTravelLocation(c)
		} else {
			address, err := nominatim.ReverseGeocode(coords.Latitude, coords.Longitude, "ru")
			if err != nil {
				log.Println(err)
			}

			return travel.IsLocationCorrect(c, *address, msg, *coords)
		}

	case states.WaitingForTravelVisitTimeStart:
		var locationData models.Location

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		startTime, err := time.Parse("02.01.2006 15:04", msg)
		if err != nil {

			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Неверно введена дата")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputVisitTimeStart(c)

		} else if startTime.After(time.Now()) {
			database.DB.Where("id=?", user.CurrentLocationId).First(&locationData)
			locationData.VisitTimeStart = startTime
			database.DB.Where("id=?", locationData.Id).Save(&locationData)

			return travel.InputVisitTimeEnd(c)
		} else {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Неверно введена дата начала посещения")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputVisitTimeStart(c)
		}

	case states.WaitingForTravelVisitTimeEnd:
		var locationData models.Location

		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		endTime, err := time.Parse("02.01.2006 15:04", msg)
		if err != nil {
			sentMsg, errSent := c.Bot().Send(c.Chat(),
				"Неверно введена дата")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return travel.InputVisitTimeEnd(c)

		} else {
			database.DB.Where("id=?", user.CurrentLocationId).First(&locationData)
			if endTime.After(locationData.VisitTimeStart) {
				locationData.VisitTimeEnd = endTime
				database.DB.Where("id=?", locationData.Id).Save(&locationData)

				return travel.LocationsMenu(c)
			} else {
				sentMsg, errSent := c.Bot().Send(c.Chat(),
					"Дата конца посещения должна быть после даты начала")

				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return travel.InputVisitTimeEnd(c)
			}
		}
	case states.WaitingForNoteName:
		var note models.Note
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		if utf8.RuneCountInString(msg) >= 2 && utf8.RuneCountInString(msg) <= 20 {
			database.DB.Where("id=?", user.CurrentNoteCreatingId).First(&note)
			if note.Id != 0 {
				note.Name = strings.TrimSpace(msg)
				database.DB.Where("id=?", note.Id).Save(&note)
				return notes.InputNoteText(c)

			} else {
				sentMsg, errSent := c.Bot().Reply(c.Message(),
					"Заметка не найдена")

				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return notes.Menu(c)
			}

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Название заметки должно быть от 2 до 20 символов, попробуй еще раз")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return notes.InputNoteName(c)
		}

	case states.WaitingForNoteText:
		var note models.Note
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()
		if utf8.RuneCountInString(msg) >= 1 && utf8.RuneCountInString(msg) <= 300 {
			database.DB.Where("id=?", user.CurrentNoteCreatingId).First(&note)
			if note.Id != 0 {
				note.Text = strings.TrimSpace(msg)
				database.DB.Where("id=?", note.Id).Save(&note)
				return notes.InputNoteFiles(c)

			} else {
				sentMsg, errSent := c.Bot().Reply(c.Message(),
					"Заметка не найдена")

				states.Sent.Mx.RLock()
				stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
				states.Sent.Map[c.Sender().ID] = stateSent
				states.Sent.Mx.RUnlock()

				if errSent != nil {
					log.Println(errSent)
				}

				return notes.Menu(c)
			}

		} else {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Текст заметки должен быть от 1 до 300 символов, попробуй еще раз")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return notes.InputNoteText(c)
		}

	case states.WaitingForTransactionAmount:
		var (
			transaction    models.Transaction
			participantIds []int64
		)
		states.Input.Mx.RLock()
		delete(states.Input.Map, userId)
		states.Input.Mx.RUnlock()

		amount, errAtoi := strconv.Atoi(msg)
		if errAtoi != nil {
			sentMsg, errSent := c.Bot().Reply(c.Message(),
				"Сумма долга должна быть числом")

			states.Sent.Mx.RLock()
			stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
			states.Sent.Map[c.Sender().ID] = stateSent
			states.Sent.Mx.RUnlock()

			if errSent != nil {
				log.Println(errSent)
			}

			return splitWise.TransactionAddMember(c, strconv.Itoa(int(user.TgId)))
		}

		database.DB.Where("id = ?", user.CurrentTransactionId).Find(&transaction)
		chosenUserId := user.CurrentTransactionUserId
		newDebt := models.Debt{
			ParticipantId: chosenUserId,
			Amount:        amount,
		}
		transaction.Participants = append(transaction.Participants, newDebt)
		database.DB.Where("id = ?", user.CurrentTransactionId).Save(&transaction)
		fmt.Printf("%+v", transaction)

		for _, debt := range transaction.Participants {
			participantIds = append(participantIds, debt.ParticipantId)
		}
		return splitWise.ChooseMember(c, participantIds)

	default:
		sentMsg, errSent := c.Bot().Reply(c.Message(),
			"Неизвестная команда")

		states.Sent.Mx.RLock()
		stateSent.SentMessagesId = append(stateSent.SentMessagesId, sentMsg.ID)
		states.Sent.Map[c.Sender().ID] = stateSent
		states.Sent.Mx.RUnlock()

		return errSent
	}
}

func checkLetters(s string) bool {
	count := 0
	for _, char := range s {
		if unicode.IsLetter(char) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}
