package main

import (
	"log"
	"time"

	bot "kmid_checker/bot"
	request "kmid_checker/request"
)

func main() {
	checkInterval := 30 * time.Minute
	timeout := 24 * time.Hour

	prevStatus := ""
	noChangeCounter := time.Duration(0)

	for {
		status, err := request.GetStatus()

		if err != nil {
			log.Printf("Error getting status: %s", err)
			time.Sleep(checkInterval)
			continue
		}

		if status == "Статус заявления: дело в обработке." {
			if status == prevStatus {
				noChangeCounter += checkInterval
				if noChangeCounter >= timeout {
					bot.SendAlert("Уведомление: статус не изменился в течении суток: " + status)
					noChangeCounter = 0
				}
			} else {
				noChangeCounter = 0
			}
		} else if status != prevStatus {
			bot.SendAlert("Уведомление: присланный статус: " + status)
		}

		prevStatus = status
		time.Sleep(checkInterval)
	}
}
