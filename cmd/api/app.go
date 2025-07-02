package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/vnam0320/tg_bot/internal"
)

type app struct {
	config internal.BotConfigure
	bot    *bot.Bot
}

func NewApp(option ...bot.Option) (*app, error) {
	cfg := internal.NewBotConfigure()

	b, err := bot.New(cfg.Token, option...)

	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	app := &app{
		config: cfg,
		bot:    b,
	}

	return app, err
}

func (app *app) Run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	app.bot.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: app.config.WebhookURL,
	})

	go func() {
		http.ListenAndServe(":"+app.config.Port, app.bot.WebhookHandler())
	}()

	log.Println("bot is running...")
	app.bot.StartWebhook(ctx)
	app.bot.DeleteWebhook(ctx, &bot.DeleteWebhookParams{
		DropPendingUpdates: true,
	})
	log.Println("bot is stopped")
}
