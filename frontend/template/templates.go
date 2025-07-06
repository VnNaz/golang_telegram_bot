package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot/models"
	"github.com/vnam0320/tg_bot/internal/model"
)

var (
	listTaskPath              = filepath.Join(os.Getenv("TEMPLATE_DIR"), "list_task.html")
	tmplListTask, errListTask = template.ParseFiles(listTaskPath)

	listMyTaskPath                = filepath.Join(os.Getenv("TEMPLATE_DIR"), "list_my_task.html")
	tmplListMyTask, errListMyTask = template.ParseFiles(listMyTaskPath)

	listOwnerTaskPath                   = filepath.Join(os.Getenv("TEMPLATE_DIR"), "list_owner_task.html")
	tmplListOwnerTask, errListOwnerTask = template.ParseFiles(listOwnerTaskPath)
)

type ListTaskData struct {
	Tasks []*model.Task
	User  *models.User
}

func ListAllTask(data *ListTaskData) (string, error) {
	var buf bytes.Buffer
	if errListTask != nil {
		return "", fmt.Errorf("can't parse text/template: %s", errListTask.Error())
	}

	errExecute := tmplListTask.Execute(&buf, data)

	if errExecute != nil {
		return "", fmt.Errorf("can't execute text/template: %s", errExecute.Error())
	}

	return buf.String(), nil
}
func ListMyTask(data *ListTaskData) (string, error) {
	var buf bytes.Buffer
	if errListMyTask != nil {
		return "", fmt.Errorf("can't parse text/template: %s", errListMyTask.Error())
	}

	errExecute := tmplListMyTask.Execute(&buf, data)

	if errExecute != nil {
		return "", fmt.Errorf("can't execute text/template: %s", errExecute.Error())
	}

	return buf.String(), nil
}

func ListOwnerTask(data *ListTaskData) (string, error) {
	var buf bytes.Buffer
	if errListOwnerTask != nil {
		return "", fmt.Errorf("can't parse text/template: %s", errListOwnerTask.Error())
	}

	errExecute := tmplListOwnerTask.Execute(&buf, data)

	if errExecute != nil {
		return "", fmt.Errorf("can't execute text/template: %s", errExecute.Error())
	}

	return buf.String(), nil
}
