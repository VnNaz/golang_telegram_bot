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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	otps := []bot.Option{
		bot.WithMiddlewares(terminalLogCommingMessage, fileLogCommingMessage),
		bot.WithDefaultHandler(getUpdates),
	}

	app, err := NewApp(otps...)

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

func terminalLogCommingMessage(next bot.HandlerFunc) bot.HandlerFunc {
	// log received message in terminal and file log
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			log.Printf("%s say: %s", update.Message.From.Username, update.Message.Text)
		}
		next(ctx, b, update)
	}
}

func fileLogCommingMessage(next bot.HandlerFunc) bot.HandlerFunc {
	//	7 = 4 (read) + 2 (write) + 1 (execute) = full access for owner

	//	5 = 4 (read) + 1 (execute) = read + enter (but not modify) for group

	//	0 = no permissions for others

	if _, err := os.Stat(os.Getenv("LOG_DIR")); err != nil {
		if os.IsNotExist(err) {
			// foulder is not exists
			err := os.Mkdir(os.Getenv("LOG_DIR"), 0750)

			if err != nil {
				log.Fatalf("failed to create log directory: %v", err)
			}
		}
	}

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			// 6 = 4 + 2 = read + write (for user)

			// 4 = read only (for group)

			// 4 = read only (for others)
			f, err := os.OpenFile(os.Getenv("LOG_FILE_PATH"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println("error opening file:", err)
				return
			}
			defer f.Close()

			logger := log.New(f, "", log.LstdFlags)
			logger.Printf("%s say: %s", update.Message.From.Username, update.Message.Text)
		}
		next(ctx, b, update)
	}
}
