package internal

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type BotConfigure struct {
	Token      string
	WebhookURL string
	Port       string
}

func getToken() string {
	return os.Getenv("BOT_TOKEN")
}

func getWebhookURL() string {
	return os.Getenv("WEBHOOK_URL")
}

func getPort() string {
	return os.Getenv("PORT")
}

func NewBotConfigure() BotConfigure {
	return BotConfigure{
		Token:      getToken(),
		WebhookURL: getWebhookURL(),
		Port:       getPort(),
	}
}
