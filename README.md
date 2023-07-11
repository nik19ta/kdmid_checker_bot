# Passport Status Checker Bot

## Description

This is a bot, written in Go, which uses cron to check the readiness status of a passport every 30 minutes. If the status doesn't change within 24 hours, the bot automatically sends a message indicating the unchanged status. For any status change, the user will be instantly notified. Additionally, the bot provides functionality to add and remove tracking of passport readiness status.

## Technologies

- Golang
- Gorm
- PostgreSQL
- Gocron
- TelegramApi

## Installation

Make sure you have Go, Gorm, and PostgreSQL installed.

1. Clone the GitHub repository:

```sh
git clone https://github.com/nik19ta/passport-status-checker-bot.git
```

2. Navigate into the project directory:

```sh
cd passport-status-checker-bot
```

3. Install the necessary dependencies:

```sh
go get
```

4. Configure parameters in .env (example in .env.example)
5. Start the bot:

```sh
make dev
```
