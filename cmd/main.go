package main

import (
	"log"
	"time"

	botPackage "kmid_checker/bot"
	corn "kmid_checker/corn"
	database "kmid_checker/pkg/database"
	env "kmid_checker/pkg/env"

	gin "github.com/gin-gonic/gin"
	gocron "github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	language "golang.org/x/text/language"

	i18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

func main() {

	// * Initializing i18n
	bundle := i18n.NewBundle(language.English)
	bundle.MustLoadMessageFile("locales/ru.json")

	localizer := i18n.NewLocalizer(bundle, "ru")

	// * Database Initialization
	db := database.InitDB()

	// * Bot Initialization
	botToken := env.Get("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	botPackage.Init(db, bot, localizer)

	// * Launch a goroutine for regular status checks (every 30 minutes)
	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(30).Minute().Do(func() {
		corn.CheckStatus(db, bot, localizer)
	})

	if err != nil {
		log.Fatalf("Could not schedule job: %v", err)
	}

	s.StartAsync()

	// * Gin Server Initialization and Launch
	r := gin.Default()
	r.GET("/example", func(ctx *gin.Context) {
		log.Println("/example")
	})
	r.Run(":3000")
}
