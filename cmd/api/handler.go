package main

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (app *app) HandlerRegisterUser() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Sucessfull registed: " + app.config.Port,
		})
	}
}

func (app *app) HandlerCommands() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		commands := make([]string, 0, len(app.cmd))

		for k := range app.cmd {
			commands = append(commands, k)
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Available commands:\n" + strings.Join(commands, "\n"),
		})
	}
}
