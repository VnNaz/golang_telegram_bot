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
	buf                       bytes.Buffer
)

type ListAllTaskData struct {
	Tasks []*model.Task
	User  *models.User
}

func ListAllTask(data *ListAllTaskData) (string, error) {

	if errListTask != nil {
		return "", fmt.Errorf("can't parse text/template: %s", errListTask.Error())
	}

	errExecute := tmplListTask.Execute(&buf, data)

	if errListTask != nil {
		return "", fmt.Errorf("can't execute text/template: %s", errExecute.Error())
	}

	return buf.String(), nil
}
