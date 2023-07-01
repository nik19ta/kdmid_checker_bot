package bot

import (
	"log"
	"strconv"

	env "kmid_checker/pkg/env"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func SendAlert(messageText string) {
	botToken := env.Get("TELEGRAM_BOT_TOKEN")
	chatIDStr := env.Get("TELEGRAM_CHAT_ID")

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Panic("Error parsing chat ID:", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("Error creating Telegram bot:", err)
	}

	msg := tgbotapi.NewMessage(chatID, messageText)
	_, err = bot.Send(msg)
	if err != nil {
		log.Panic("Error sending message:", err)
	}
}
