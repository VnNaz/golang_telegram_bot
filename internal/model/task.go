package model

import (
	"github.com/go-telegram/bot/models"
)

type Task struct {
	Id          int64
	Assignee    *models.User
	Description string
	Onwer       *models.User
}
