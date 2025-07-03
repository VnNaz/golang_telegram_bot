package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/go-telegram/bot/models"
	"github.com/vnam0320/tg_bot/internal/model"
)

var TaskIsNotExist = errors.New("Task with given id is not exist")
var NotYourTask = errors.New("Current task is not belong to you")

type Taskstorage interface {
	Create(ctx context.Context, u *models.User, description string) (*model.Task, error)
	Unassign(ctx context.Context, u *models.User, taskId int64) (*model.Task, error)
	Assign(ctx context.Context, u *models.User, taskId int64) (*model.Task, error)
	GetByOnwer(ctx context.Context, u *models.User) ([]*model.Task, error)
	GetByAssignee(ctx context.Context, u *models.User) ([]*model.Task, error)
	GetById(ctx context.Context, id int64) (*model.Task, error)
	GetAll(ctx context.Context) ([]*model.Task, error)
	Resolve(ctx context.Context, u *models.User, taskId int64) (*model.Task, error)
}

type InMemoryTaskStorage struct {
	tasks    map[int64]*model.Task
	mu       sync.RWMutex
	sequence int64
}

func NewInMemoryTaskStorage() *InMemoryTaskStorage {
	store := InMemoryTaskStorage{
		mu:       sync.RWMutex{},
		tasks:    make(map[int64]*model.Task),
		sequence: 0,
	}
	return &store
}

func (store *InMemoryTaskStorage) nextId() int64 {
	// this function should called within mutex
	store.sequence++
	return store.sequence
}

func (store *InMemoryTaskStorage) Create(ctx context.Context, u *models.User, description string) (*model.Task, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	task := &model.Task{
		Description: description,
		Onwer:       u,
		Assignee:    nil,
		Id:          store.nextId(),
	}

	store.tasks[task.Id] = task
	return task, nil
}

func (store *InMemoryTaskStorage) Assign(ctx context.Context, u *models.User, taskId int64) (*model.Task, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if task, ok := store.tasks[taskId]; !ok {
		return nil, TaskIsNotExist
	} else {
		task.Assignee = u
		return task, nil
	}
}

func (store *InMemoryTaskStorage) Unassign(ctx context.Context, u *models.User, taskId int64) (*model.Task, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if task, ok := store.tasks[taskId]; !ok {
		return nil, TaskIsNotExist
	} else if task.Assignee.ID != u.ID {
		return nil, NotYourTask
	} else {
		task.Assignee = nil
		return task, nil
	}
}

func (store *InMemoryTaskStorage) GetByOnwer(ctx context.Context, u *models.User) ([]*model.Task, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []*model.Task
	for _, v := range store.tasks {
		if v.Onwer.ID == u.ID {
			result = append(result, v)
		}
	}

	return result, nil
}

func (store *InMemoryTaskStorage) GetByAssignee(ctx context.Context, u *models.User) ([]*model.Task, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []*model.Task
	for _, v := range store.tasks {
		if v.Assignee.ID == u.ID {
			result = append(result, v)
		}
	}

	return result, nil
}

func (store *InMemoryTaskStorage) GetById(ctx context.Context, id int64) (*model.Task, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if task, ok := store.tasks[id]; !ok {
		return nil, TaskIsNotExist
	} else {
		return task, nil
	}
}

func (store *InMemoryTaskStorage) GetAll(ctx context.Context) ([]*model.Task, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []*model.Task
	for _, v := range store.tasks {
		result = append(result, v)
	}

	return result, nil
}

func (store *InMemoryTaskStorage) Resolve(ctx context.Context, u *models.User, taskId int64) (*model.Task, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if task, ok := store.tasks[taskId]; !ok {
		return nil, TaskIsNotExist
	} else if task.Assignee.ID != u.ID {
		return nil, NotYourTask
	} else {
		delete(store.tasks, taskId)
		return task, nil
	}
}
