package main

import (
	"backend-qushedo/database"
	"backend-qushedo/openTripMap"
	"backend-qushedo/routes"
	"backend-qushedo/states"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	tb "gopkg.in/telebot.v3"
)

var _ = godotenv.Load() // It's very unsafe, but okay

var (
	token     = os.Getenv("TOKEN")
	otmApiKey = os.Getenv("OTM_API_KEY")
)

func main() {
	database.Connect()
	states.MakeInputMap()
	states.MakeSentMap()

	pref := tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	openTripMap.NewClient(otmApiKey)

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	routes.Setup(b)
	log.Println("Бот запущен")
	b.Start()
}
