package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var logger = new(log.Logger)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	f := initLogStream(os.Getenv("LOG_DIR"), os.Getenv("LOG_FILE_PATH"))
	defer f.Close()

	// combine log cli and file stream
	multiwriter := io.MultiWriter(os.Stdout, f)
	// init logger
	logger = log.New(multiwriter, "", log.LstdFlags)

	otps := []bot.Option{
		bot.WithMiddlewares(loggingIncommingMessage(logger)),
		bot.WithDefaultHandler(getUpdates),
	}

	app, err := NewApp(otps...)

	if err != nil {
		logger.Fatal(err)
	}

	app.Run(ctx, cancel)
}

func getUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}

func initLogStream(dir, file string) io.WriteCloser {
	//	7 = 4 (read) + 2 (write) + 1 (execute) = full access for owner
	//	5 = 4 (read) + 1 (execute) = read + enter (but not modify) for group
	//	0 = no permissions for others
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// foulder is not exists
			err := os.Mkdir(dir, 0750)

			if err != nil {
				log.Fatalf("failed to create log directory: %v", err)
			}
		}
	}

	// 6 = 4 + 2 = read + write (for user)
	// 4 = read only (for group)
	// 4 = read only (for others)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("error opening file:", err)
	}

	return f
}

func loggingIncommingMessage(log *log.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message != nil {
				log.Printf("%s say: %s", update.Message.From.Username, update.Message.Text)
			}
			next(ctx, b, update)
		}
	}
}
