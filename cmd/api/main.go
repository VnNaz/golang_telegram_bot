package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// TODO: write log into file for each message
	// TODO: wrap the log object to add more information like time, level, etc.
	// TODO: automatic rerun
	// go build -mod=vendor -o .\bin\main.exe .\cmd\api\
	// .\bin\main.exe
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	app, err := NewApp(getUpdates)

	if err != nil {
		log.Fatal(err)
	}

	app.Run(ctx, cancel)
}

func getUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}
