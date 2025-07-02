package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (app *app) HandlerRegister() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		err := app.store.Users.Register(ctx, update.Message.From)
		text := "Successfully registed user: " + update.Message.From.Username
		if err != nil {
			text = fmt.Sprintf("failed add user: %s", err.Error())
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
		})
	}
}

func (app *app) HandlerUnregister() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		err := app.store.Users.Unregister(ctx, update.Message.From)
		text := "Successfully unregistered user: " + update.Message.From.Username
		if err != nil {
			text = fmt.Sprintf("failed unregistered user: %s", err.Error())
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
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
