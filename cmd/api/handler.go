package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	templates "github.com/vnam0320/tg_bot/frontend/template"
	"github.com/vnam0320/tg_bot/internal/storage"
)

type HandlerConfiguration struct {
	cmd       string
	re        *regexp.Regexp
	matchType bot.MatchType
	handler   bot.HandlerFunc
	NotShowed bool
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
			// TODO: if in database this can be skipped
			sort.Slice(tasks, func(i, j int) bool {
				return tasks[i].Id < tasks[j].Id
			})
			// example: 1. написать бота by @ivanov
			text, err := templates.ListAllTask(&templates.ListTaskData{
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
		cmd:       "/assign_<id>",
		re:        re,
		NotShowed: true,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			matches := re.FindStringSubmatch(update.Message.Text)
			if len(matches) > 1 {
				// regexp catch only id is number, so this is alway sucessful, so skip the error
				taskId, _ := strconv.Atoi(matches[1])
				// check if task is exists
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

func (app *app) HandlerUnassignTask() *HandlerConfiguration {
	re := regexp.MustCompile(`^/unassign_(\d+)$`)
	return &HandlerConfiguration{
		cmd:       "/unassign_<id>",
		re:        re,
		NotShowed: true,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			matches := re.FindStringSubmatch(update.Message.Text)
			if len(matches) > 1 {
				// regexp catch only id is number, so this is alway sucessful, so skip the error
				taskId, _ := strconv.Atoi(matches[1])

				task, err := app.store.Tasks.Unassign(ctx, update.Message.From, int64(taskId))

				switch {
				case errors.Is(err, storage.NotYourTask):
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   "Задача не на вас",
					})
				case errors.Is(err, storage.TaskIsNotExist):
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача с идентификатором %d не сушествует", taskId),
					})
				case err != nil:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Не могу отменить вам эту задачу: %s", err.Error()),
					})
				default:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   "Принято",
					})
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: task.Onwer.ID,
						Text:   fmt.Sprintf("Задача \"%s\" осталась без исполнителя", task.Description),
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

func (app *app) HandlerResolveTask() *HandlerConfiguration {
	re := regexp.MustCompile(`^/resolve_(\d+)$`)
	return &HandlerConfiguration{
		cmd:       "/resolve_<id>",
		re:        re,
		NotShowed: true,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			matches := re.FindStringSubmatch(update.Message.Text)
			if len(matches) > 1 {
				// regexp catch only id is number, so this is alway sucessful, so skip the error
				taskId, _ := strconv.Atoi(matches[1])

				task, err := app.store.Tasks.Resolve(ctx, update.Message.From, int64(taskId))

				switch {
				case errors.Is(err, storage.NotYourTask):
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   "Задача не на вас",
					})
				case errors.Is(err, storage.TaskIsNotExist):
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача с идентификатором %d не сушествует", taskId),
					})
				case err != nil:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Не могу удалить вам эту задачу: %s", err.Error()),
					})
				case update.Message.Chat.ID == task.Onwer.ID:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" выполнена", task.Description),
					})
				default:
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   fmt.Sprintf("Задача \"%s\" выполнена", task.Description),
					})
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: task.Onwer.ID,
						Text:   fmt.Sprintf("Задача \"%s\" выполнена @%s", task.Description, update.Message.From.Username),
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

func (app *app) HandlerShowMyTask() *HandlerConfiguration {
	return &HandlerConfiguration{
		cmd:       "/my",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			tasks, err := app.store.Tasks.GetByAssignee(ctx, update.Message.From)
			switch {
			case err != nil:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   fmt.Sprintf("Невозможно получить ваши задачи: %s", err.Error()),
				})
			case len(tasks) == 0:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "У вас нет задачи",
				})
			default:
				text, err := templates.ListMyTask(&templates.ListTaskData{
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
			}
		},
	}
}

func (app *app) HandlerShowOwnerTask() *HandlerConfiguration {
	return &HandlerConfiguration{
		cmd:       "/owner",
		matchType: bot.MatchTypeExact,
		handler: func(ctx context.Context, b *bot.Bot, update *models.Update) {
			tasks, err := app.store.Tasks.GetByOnwer(ctx, update.Message.From)
			switch {
			case err != nil:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   fmt.Sprintf("Невозможно получить ваши задачи: %s", err.Error()),
				})
			case len(tasks) == 0:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "У вас нет задачи",
				})
			default:
				text, err := templates.ListOwnerTask(&templates.ListTaskData{
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
			}
		},
	}
}
