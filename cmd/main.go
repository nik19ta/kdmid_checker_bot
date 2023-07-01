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
						ApplicationNumber: "0",
						CityID:            0,
						PassportType:      "5",
					}

					db.Create(&request)

					bot.Send(newmsg)
				} else if buttonData == "10" {
					newmsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пожалуйста укажите номер заявления")

					request := models.Request{
						ID:                uuid.New(),
						UserID:            update.CallbackQuery.Message.Chat.ID,
						ApplicationNumber: "0",
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

				var request models.Request
				if err := db.Where("user_id = ?", update.Message.Chat.ID).First(&request).Error; err != nil {
					if err != gorm.ErrRecordNotFound {
						log.Println("Database error:", err)
					}
				} else {
					log.Println("Found request:", request)
				}

				switch update.Message.Command() {
				case "start":
					if request.UserID != 0 {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже есть заявление которые вы отслеживаете с номером: "+request.ApplicationNumber)
						bot.Send(errorMsg)
					} else {
						msg.Text = "Выберите срок действия паспорта:"
						msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("5 лет", "5"),
								tgbotapi.NewInlineKeyboardButtonData("10 лет", "10"),
							),
						)
					}
				case "remove":
					if request.UserID != 0 {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша заявка была удалена, и больше не будет отслеживаться.")
						bot.Send(errorMsg)
						db.Delete(&request)
					} else {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет активной заявки")
						bot.Send(errorMsg)
					}
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
				if request.PassportType == "5" && request.ApplicationNumber == "0" {

					status, err := requestPackege.GetStatusFiveYears(update.Message.Text)

					if err != nil {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получение статуса от сервиса passportzu.kdmid.ru, пожалуйста, напишите @nik19ta что бы решить проблему")
						bot.Send(finishMsg)
					}

					log.Println(status)

					if status == "Заявление с таким номером не было сохранено на сайте." {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заявление с таким номером не было сохранено на сайте.\nВозможно вы указали не верный номер, попробуйте сново")
						bot.Send(finishMsg)
					} else if status == "Статус заявления: паспорт готов." {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш паспорт уже готов.")
						bot.Send(finishMsg)
						db.Delete(&request)
					} else {
						// ! Обновляю в базе данных
						db.Model(&request).Update("application_number", update.Message.Text)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш номер заявления сохранён, мы будем проверять статус каждые пол часа, если в течении суток статус не меняеться, мы отсылаем вам текущий статус. \n\nКак только статус заявки измениться, мы пришлём вам уведомление, ваш текущий статус \""+status+"\""+"\n\nПожалуйста, не отключайте уведомления что бы вы могли сразу узнать готовность вашего документа")
						bot.Send(finishMsg)
					}
				}

				// ! Если паспорт на 10 лет
				if request.PassportType == "10" {

					if request.ApplicationNumber == "0" {
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
