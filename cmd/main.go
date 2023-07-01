package main

import (
	"log"

	"kmid_checker/models"
	database "kmid_checker/pkg/database"
	env "kmid_checker/pkg/env"

	requestPackege "kmid_checker/request"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	// Инициализация базы данных
	db := database.InitDB()

	// Инициализация бота
	botToken := env.Get("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// Запуск горутины для обработки сообщений бота
	go func() {
		for update := range updates {

			// * Если кнопки
			if update.CallbackQuery != nil {
				buttonData := update.CallbackQuery.Data

				if buttonData == "5" {
					newmsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста укажите номер заявления")

					request := models.Request{
						ID:                uuid.New(),
						UserID:            update.CallbackQuery.Message.Chat.ID,
						ApplicationNumber: 0,
						CityID:            0,
						PassportType:      "5",
					}

					db.Create(&request)

					bot.Send(newmsg)
				} else if buttonData == "10" {
					log.Println("Бот на 10 лет")

					newmsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста укажите номер заявления")

					request := models.Request{
						ID:                uuid.New(),
						UserID:            update.CallbackQuery.Message.Chat.ID,
						ApplicationNumber: 0,
						CityID:            0,
						PassportType:      "10",
					}

					db.Create(&request)

					bot.Send(newmsg)
				}
				continue
			}

			// * Если сообщение пустое
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			// * Если у нас команда
			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				log.Println(update.Message)

				switch update.Message.Command() {
				case "start":
					msg.Text = "Выберите срок действия паспорта:"
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("5 лет", "5"),
							tgbotapi.NewInlineKeyboardButtonData("10 лет", "10"),
						),
					)
				default:
					msg.Text = "Я не знаю эту команду"
				}
				bot.Send(msg)

				continue
			}

			if update.Message != nil {
				// * Проверяем сохарнён ли у нас такой пользователь
				var request models.Request
				if err := db.Where("user_id = ?", update.Message.Chat.ID).First(&request).Error; err != nil {
					if err != gorm.ErrRecordNotFound {
						log.Println("Database error:", err)
					}
				} else {
					log.Println("Found request:", request)
				}

				// ! Если паспорт на 5 лет
				if request.PassportType == "5" && request.ApplicationNumber == 0 {
					db.Model(&request).Update("application_number", update.Message.Text)

					finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш номер заявления сохранён, мы будем проверять статус каждые пол часа, если в течении суток статус не меняеться, мы отсылаем вам текущий статус. \n\nКак только статус заявки измениться, мы пришлём вам уведомление")
					bot.Send(finishMsg)
				}

				// ! Если паспорт на 10 лет
				if request.PassportType == "10" {

					if request.ApplicationNumber == 0 {
						db.Model(&request).Update("application_number", update.Message.Text)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста укажите город в котором вы подавали заявления (кириллицей). \nНапример 'Астана'")
						bot.Send(finishMsg)
					} else if request.CityID == 0 {

						cityID, err := requestPackege.GetCityIdByName(update.Message.Text)

						if err != nil {
							errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой город не найден :(. Проверьте может вы написали его не правильно или это его старое название.")
							bot.Send(errorMsg)
						} else {
							db.Model(&request).Update("city_id", cityID)

							finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш номер заявления сохранён, мы будем проверять статус каждые пол часа, если в течении суток статус не меняеться, мы отсылаем вам текущий статус. \n\nКак только статус заявки измениться, мы пришлём вам уведомление")
							bot.Send(finishMsg)
						}
					}
				}
			}
		}
	}()

	// Инициализация и запуск сервера GIN
	r := gin.Default()
	// Здесь добавьте свои маршруты
	r.GET("/example", func(ctx *gin.Context) {
		log.Println("/example")
	})
	r.Run(":3000")
}
