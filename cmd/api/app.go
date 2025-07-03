package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-telegram/bot"
	config "github.com/vnam0320/tg_bot/internal/config"
	"github.com/vnam0320/tg_bot/internal/storage"
)

type app struct {
	config config.BotConfigure
	bot    *bot.Bot
	cmd    map[string]interface{}
	store  *storage.Store
}

func NewApp(option ...bot.Option) (*app, error) {
	cfg := config.NewBotConfigure()

	b, err := bot.New(cfg.Token, option...)

	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	app := &app{
		config: cfg,
		bot:    b,
		cmd:    make(map[string]interface{}),
		store:  storage.NewStore(),
	}
	// to do reflex to register all function with handler
	app.MountHandler()

	return app, err
}

func (app *app) MountHandler() {
	// reflect register handler
	t := reflect.TypeOf(app)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)

		if !strings.HasPrefix(method.Name, "Handler") {
			continue
		}

		log.Printf("calling method %s\n", method.Name)

		receiver := reflect.ValueOf(app)
		results := method.Func.Call([]reflect.Value{receiver})

		if len(results) != 1 {
			log.Printf("skip handler %s because not enough return value \n", method.Name)
			continue
		}

		config, ok := results[0].Interface().(*HandlerConfiguration)

		if !ok {
			log.Printf("skip handler %s because return value have unexpected types \n", method.Name)
			continue
		}
		// if regexp is passed
		if config.re != nil {
			app.bot.RegisterHandlerRegexp(bot.HandlerTypeMessageText, config.re, config.handler)
		} else {
			app.bot.RegisterHandler(bot.HandlerTypeMessageText, config.cmd, config.matchType, config.handler)
		}

		app.cmd[config.cmd] = nil
		log.Printf("registed handler %s with command %s \n", method.Name, config.cmd)
	}
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
