package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	templates "github.com/vnam0320/tg_bot/frontend/template"
)

type HandlerConfiguration struct {
	cmd       string
	re        *regexp.Regexp
	matchType bot.MatchType
	handler   bot.HandlerFunc
}

func (app *app) HandlerRegisterUser() *HandlerConfiguration {
	return &HandlerConfiguration{
		cmd:       "/register",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			err := app.store.Users.Register(ctx, update.Message.From)
			text := "Пользователь успешно зареган: " + update.Message.From.Username
			if err != nil {
				text = fmt.Sprintf("Нельзя загерать пользователя: %s", err.Error())
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
			})
		},
	}
}

func (app *app) HandlerUnregisterUser() *HandlerConfiguration {
	return &HandlerConfiguration{
		cmd:       "/unregister",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			err := app.store.Users.Unregister(ctx, update.Message.From)
			text := "Пользователь успешно раззареган: " + update.Message.From.Username
			if err != nil {
				text = fmt.Sprintf("Нельзя раззагерать пользователя: %s", err.Error())
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
			})
		},
	}
}

func (app *app) HandlerShowCommands() *HandlerConfiguration {
	return &HandlerConfiguration{
		cmd:       "/commands",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			commands := make([]string, 0, len(app.cmd))
			for k := range app.cmd {
				commands = append(commands, k)
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Все доступные команды:\n" + strings.Join(commands, "\n"),
			})
		},
	}
}

func (app *app) HandlerListAllTasks() *HandlerConfiguration {

	return &HandlerConfiguration{
		cmd:       "/tasks",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {

			tasks, err := app.store.Tasks.GetAll(ctx)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   fmt.Sprintf("Не возможно получить задачи: %s", err.Error()),
				})
			} else if len(tasks) == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Нет задач",
				})
			}
			// example: 1. написать бота by @ivanov
			text, err := templates.ListAllTask(&templates.ListAllTaskData{
				Tasks: tasks,
				User:  update.Message.From,
			})
			if err != nil {
				logger.Println(err.Error())
				text = err.Error()
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   text,
			})
		},
	}
}

// /new XXX YYY ZZZ - создаёт новую задачу
func (app *app) HandlerCreateNewTask() *HandlerConfiguration {
	re := regexp.MustCompile(`^/new\s+(.+)$`)
	return &HandlerConfiguration{
		cmd: "/new вводите задачу",
		re:  regexp.MustCompile(`^/new(.*)$`), // catch all start with new // TODO: hard to understand, should be other handler to catch this case -> command
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			matches := re.FindStringSubmatch(update.Message.Text)
			if len(matches) > 1 {
				task, err := app.store.Tasks.Create(ctx, update.Message.From, matches[1])
				if err != nil {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   "Не могу добавить задачу в список",
					})
				} else {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" создана, id=%d", task.Description, task.Id),
					})
				}
			} else {
				// TODO: should put to other handlers
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Неверная команда, прочекайте здесь /commands",
				})
			}
		},
	}
}

func (app *app) HandlerAssignTask() *HandlerConfiguration {
	re := regexp.MustCompile(`^/assign_(\d+)$`)
	return &HandlerConfiguration{
		cmd: "/assign_<id>",
		re:  re,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			matches := re.FindStringSubmatch(update.Message.Text)
			if len(matches) > 1 {
				// regexp catch only id is number, so this is alway sucessful, so skip the error
				taskId, _ := strconv.Atoi(matches[1])

				task, err := app.store.Tasks.GetById(ctx, int64(taskId))

				switch {
				case err != nil:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Не могу назначить вам эту задачу: %s", err.Error()),
					})
				case task.Assignee == nil:
					app.store.Tasks.Assign(ctx, update.Message.From, int64(taskId))
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" назначена на вас", task.Description),
					})
				case task.Assignee.ID == update.Message.From.ID:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" уже назначена на вас", task.Description),
					})
				case task.Assignee.ID != update.Message.From.ID:
					oldAssignee := task.Assignee
					app.store.Tasks.Assign(ctx, update.Message.From, int64(taskId))
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" назначена на вас", task.Description),
					})

					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: oldAssignee.ID,
						Text:   fmt.Sprintf("Задача \"%s\" назначена на @%s", task.Description, update.Message.From.Username),
					})
				}
			} else {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Неверная команда, прочекайте здесь /commands",
				})
			}
		},
	}
}
