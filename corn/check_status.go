package corn

import (
	models "kmid_checker/models"
	"log"

	requestPackege "kmid_checker/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	i18n "github.com/nicksnyder/go-i18n/v2/i18n"
	gorm "gorm.io/gorm"
)

func CheckStatus(db *gorm.DB, bot *tgbotapi.BotAPI, localizer *i18n.Localizer) {
	var requests []models.Request

	result := db.Find(&requests)

	if result.Error != nil {
		log.Println("Error", result.Error)
	}

	for _, request := range requests {

		// * Check if the application is empty, then skipping
		// * If the passport is valid for 5 years:
		// * - ApplicationNumber is empty, then skipping
		// * If the passport is valid for 10 years:
		// * - ApplicationNumber is empty, then skipping
		// * - CityID is empty, then skipping
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

		// * Check if the status has changed for passports valid for 5 years.
		if request.PassportType == "5" {
			status, err := requestPackege.GetStatusFiveYears(request.ApplicationNumber)

			if err != nil {
				log.Println(err)
			}

			// * Check if the status has changed
			if request.Status != status {
				// * If the status has changed, we notify the user about it
				// * and update the status of the current application in the database
				msg := tgbotapi.NewMessage(request.UserID, localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "application_status_changed",
						Other: "The status of your application has been updated to: {{.Status}}",
					},
					TemplateData: map[string]interface{}{
						"Status": status,
					},
				}))
				bot.Send(msg)
				db.Model(&request).Update("status", status)
			} else {
				// * If the status has not changed, we check how many checks have been made
				// * If the number of checks is greater than 48 (24 hours)
				// * Then we send a message to the user that the status has not changed
				// * and reset the check counter
				if request.NumberChecksToday >= 48 {
					msg := tgbotapi.NewMessage(request.UserID, localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "application_status_not_changed",
							Other: "The status of your application has not changed in the last 24 hours:\n\n{{.Status}}",
						},
						TemplateData: map[string]interface{}{
							"Status": status,
						},
					}))
					bot.Send(msg)
					db.Model(&request).Update("number_checks_today", 0)
				} else {
					// * If the previous count is less than 48 checks, then we add 1 to the current count
					db.Model(&request).Update("number_checks_today", request.NumberChecksToday+1)
				}

			}
		}

		// * For passports valid for 10 years
		if request.PassportType == "10" {
			// ! Checking the status of passports valid for 10 years
		}
	}

}
