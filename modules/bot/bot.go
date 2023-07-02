package bot

import (
	"kmid_checker/models"
	"log"

	requestPackage "kmid_checker/modules/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"
)

func Init(db *gorm.DB, bot *tgbotapi.BotAPI, localizer *i18n.Localizer) {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Println(err)
	}

	// * Launch a goroutine for processing bot messages
	go func() {
		for update := range updates {

			// * If buttons
			if update.CallbackQuery != nil {
				buttonData := update.CallbackQuery.Data

				if buttonData == "5" {
					newmsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
						localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "please_provide_application_number",
								Other: "Please provide application number",
							},
						}))

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
					newmsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
						localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "please_provide_application_number",
								Other: "Please provide application number",
							},
						}))

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

			// * If the message is empty
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			// * If we have a command
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
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "your_application_is_being_checked",
									Other: "Your application with the number {{.ApplicationNumber}} is checked every 30 minutes:\n\n{{.Status}}",
								},
								TemplateData: map[string]interface{}{
									"ApplicationNumber": request.ApplicationNumber,
									"Status":            request.Status,
								},
							}))
						bot.Send(errorMsg)
					} else {
						msg.Text = localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "please_select_passport_validity_period",
								Other: "Please select the passport validity period.",
							},
						})

						msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData(
									localizer.MustLocalize(&i18n.LocalizeConfig{
										DefaultMessage: &i18n.Message{
											ID:    "five_years",
											Other: "Five years",
										},
									}),
									"5",
								),
								tgbotapi.NewInlineKeyboardButtonData(
									localizer.MustLocalize(&i18n.LocalizeConfig{
										DefaultMessage: &i18n.Message{
											ID:    "ten_years",
											Other: "Ten years",
										},
									}),
									"10",
								),
							),
						)
					}
				case "remove":
					if request.UserID != 0 {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "your_application_was_deleted",
									Other: "Your application (#{{.ApplicationNumber}}) has been deleted and will no longer be tracked.",
								},
								TemplateData: map[string]interface{}{
									"ApplicationNumber": request.ApplicationNumber,
								},
							}))
						bot.Send(errorMsg)
						db.Delete(&request)
					} else {
						errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "no_active_application",
									Other: "You don't have an active application.",
								},
							}))
						bot.Send(errorMsg)
					}
				default:
					msg.Text = localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "unknown_command",
							Other: "Unknown command",
						},
					})
				}
				bot.Send(msg)
				continue
			}

			if update.Message != nil {
				// * Check if such user is saved
				var request models.Request
				if err := db.Where("user_id = ?", update.Message.Chat.ID).First(&request).Error; err != nil {
					if err != gorm.ErrRecordNotFound {
						log.Println("Database error:", err)
					}
				}

				// * If the passport is valid for 5 years
				if request.PassportType == "5" && request.ApplicationNumber == "0" {

					status, err := requestPackage.GetStatusFiveYears(update.Message.Text)

					if err != nil {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "error_getting_status",
									Other: "An error occurred while retrieving the status. Please contact @nik19ta to resolve the issue.",
								},
							}))

						bot.Send(finishMsg)
					}

					if status == "Заявление с таким номером не было сохранено на сайте." {

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "no_saved_application",
									Other: "The application with that number has not been saved on the website. \nPerhaps you entered an incorrect number. Please try again.",
								},
							}))
						bot.Send(finishMsg)
					} else if status == "Статус заявления: паспорт готов." {
						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID, localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "your_document_is_ready",
								Other: "Your document is already ready.",
							},
						}))

						bot.Send(finishMsg)
						db.Delete(&request)
					} else {
						// * Update in the database
						// * - Set the application number
						// * - Set the current application status
						// * - Set the number of checks for today (1)
						db.Model(&request).Update("application_number", update.Message.Text)
						db.Model(&request).Update("status", status)
						db.Model(&request).Update("number_checks_today", 1)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "application_saved",
									Other: "Your application number has been saved, we will check the status every half an hour, if the status does not change within a day, we will send you the current status. \n\nAs soon as the application status changes, we will send you a notification, your current status is \"{{.Status}}\"\n\nPlease do not turn off notifications so you can immediately find out the readiness of your document.",
								},
								TemplateData: map[string]interface{}{
									"Status": request.Status,
								},
							}))
						bot.Send(finishMsg)
					}
				}

				// * If the passport is valid for 10 years
				if request.PassportType == "10" {

					if request.ApplicationNumber == "0" {
						db.Model(&request).Update("application_number", update.Message.Text)

						finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							localizer.MustLocalize(&i18n.LocalizeConfig{
								DefaultMessage: &i18n.Message{
									ID:    "please_specify_the_city_where_you_submitted_the_application",
									Other: "Your application number has been saved, we will check the status every half an hour, if the status does not change within a day, we will send you the current status. \n\nAs soon as the application status changes, we will send you a notification, your current status is \"{{.Status}}\"\n\nPlease do not turn off notifications so you can immediately find out the readiness of your document.",
								},
								TemplateData: map[string]interface{}{
									"Status": request.Status,
								},
							}))
						bot.Send(finishMsg)
					} else if request.CityID == 0 {

						cityID, err := requestPackage.GetCityIdByName(update.Message.Text)

						if err != nil {
							errorMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
								localizer.MustLocalize(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										ID:    "the_city_was_not_found",
										Other: "The city was not found. Please double-check if you spelled it correctly or if it may be known by a different name.",
									},
									TemplateData: map[string]interface{}{
										"Status": request.Status,
									},
								}))
							bot.Send(errorMsg)
						} else {
							db.Model(&request).Update("city_id", cityID)

							finishMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
								localizer.MustLocalize(&i18n.LocalizeConfig{
									DefaultMessage: &i18n.Message{
										ID:    "application_saved",
										Other: "Your application number has been saved, we will check the status every half an hour, if the status does not change within a day, we will send you the current status. \n\nAs soon as the application status changes, we will send you a notification, your current status is \"{{.Status}}\"\n\nPlease do not turn off notifications so you can immediately find out the readiness of your document.",
									},
									TemplateData: map[string]interface{}{
										"Status": request.Status,
									},
								}))
							bot.Send(finishMsg)
						}
					}
				}
			}
		}
	}()
}
