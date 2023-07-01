package main

import (
	"log"
	"time"

	models "kmid_checker/models"
	database "kmid_checker/pkg/database"
	env "kmid_checker/pkg/env"

	requestPackege "kmid_checker/request"

	gin "github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	uuid "github.com/google/uuid"
	gorm "gorm.io/gorm"

	gocron "github.com/go-co-op/gocron"
)

func checkStatus(db *gorm.DB, bot *tgbotapi.BotAPI) {
	var requests []models.Request

	result := db.Find(&requests)

	if result.Error != nil {
		log.Println("Error", result.Error)
	}

	for _, request := range requests {

		// * Проверяем, если заявка пустая, то пропускаем
		// * Если паспорт на 5 лет и:
		// * - ApplicationNumber пустой, пропускаем
		// * Если паспорт на 10 лет и:
		// * - ApplicationNumber пустой, пропускаем
		// * - CityID пустой, пропускаем
		if request.UserID == 0 {
			continue
		}

		if request.PassportType == "5" {
			if request.ApplicationNumber == "0" {
				continue
			}
		} else if request.PassportType == "10" {
			if request.CityID == 0 || request.ApplicationNumber == "0" {
				continue
			}
		} else {
			continue
		}

		// * Проверяем изменился ли статус для пасппортов 5 лет.
		if request.PassportType == "5" {
			status, err := requestPackege.GetStatusFiveYears(request.ApplicationNumber)

			if err != nil {
				log.Println(err)
			}

			// * Проверяем изменился ли статус
			if request.Status != status {
				// * Если статус изменился то уведомляем пользователя об этом
				// * И изменяем в базе данных статус текущей заявки
				msg := tgbotapi.NewMessage(request.UserID, "Статус заявления изменился:\n\n"+status)
				bot.Send(msg)
				db.Model(&request).Update("status", status)
			} else {
				// * Если статус не изменился то проверяем, сколько было проверок
				// * Если колличество проверок больше чем 48 (сутки)
				// * То отсылаем сообщение пользователю что статус не изменился
				// * И обнуляем счётчик проверок
				if request.NumberChecksToday >= 48 {
					msg := tgbotapi.NewMessage(request.UserID, "Статус заявления за последние 24 часа не изменился:\n\n"+status)
					bot.Send(msg)
					db.Model(&request).Update("number_checks_today", 0)
				} else {
					// * Если прошлом меньше чем 48 проверок то добовляем к текущем + 1
					db.Model(&request).Update("number_checks_today", request.NumberChecksToday+1)
				}

			}
		}

		// * Для паспортов 10 лет
		if request.PassportType == "10" {
			// ! Проверяем статус паспорта на 10 лет,
		}
	}

}

func main() {
	// * Инициализация базы данных
	db := database.InitDB()

	// * Инициализация бота
	botToken := env.Get("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// * Запуск горутины для регулярной проверки статусов ( каждые 30 минут )
	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(30).Minute().Do(func() {
		checkStatus(db, bot)
	})

	if err != nil {
		log.Fatalf("Could not schedule job: %v", err)
	}

	s.StartAsync()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// * Запуск горутины для обработки сообщений бота
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
						NumberChecksToday: 0,
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
						NumberChecksToday: 0,
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
				}

				switch update.Message.Command() {
				case "start":
					if request.UserID != 0 {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваше заявление с номером "+request.ApplicationNumber+" проверяеться каждые 30 минут:\n\n"+request.Status)
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
				}

				// * Если паспорт на 5 лет
				if request.PassportType == "5" && request.ApplicationNumber == "0" {

					status, err := requestPackege.GetStatusFiveYears(update.Message.Text)

					if err != nil {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получение статуса от сервиса passportzu.kdmid.ru, пожалуйста, напишите @nik19ta что бы решить проблему")
						bot.Send(finishMsg)
					}

					if status == "Заявление с таким номером не было сохранено на сайте." {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заявление с таким номером не было сохранено на сайте.\nВозможно вы указали не верный номер, попробуйте сново")
						bot.Send(finishMsg)
					} else if status == "Статус заявления: паспорт готов." {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш паспорт уже готов.")
						bot.Send(finishMsg)
						db.Delete(&request)
					} else {
						// * Обновляю в базе данных
						// * - Устанавливаем номер заявления
						// * - Устанавливаем статус заявления на данный момент
						// * - Устанавливаем колличество проверок на сегодня (1)
						db.Model(&request).Update("application_number", update.Message.Text)
						db.Model(&request).Update("status", status)
						db.Model(&request).Update("number_checks_today", 1)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш номер заявления сохранён, мы будем проверять статус каждые пол часа, если в течении суток статус не меняеться, мы отсылаем вам текущий статус. \n\nКак только статус заявки измениться, мы пришлём вам уведомление, ваш текущий статус \""+status+"\""+"\n\nПожалуйста, не отключайте уведомления что бы вы могли сразу узнать готовность вашего документа")
						bot.Send(finishMsg)
					}
				}

				// * Если паспорт на 10 лет
				if request.PassportType == "10" {

					if request.ApplicationNumber == "0" {
						db.Model(&request).Update("application_number", update.Message.Text)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста укажите город в котором вы подавали заявления (кириллицей). \nНапример 'Астана'")
						bot.Send(finishMsg)
					} else if request.CityID == 0 {

						cityID, err := requestPackege.GetCityIdByName(update.Message.Text)

						if err != nil {
							errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой город не найден. Проверьте может вы написали его не правильно или это его старое название.")
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
