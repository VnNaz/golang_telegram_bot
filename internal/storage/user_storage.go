package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-telegram/bot/models"
)

type UserStorage interface {
	Register(ctx context.Context, u *models.User) error
	Unregister(ctx context.Context, u *models.User) error
}

type InMemoryUserStorage struct {
	users map[int64]*models.User
	mu    sync.RWMutex
}

func (store *InMemoryUserStorage) Register(ctx context.Context, u *models.User) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.users[u.ID]; ok {
		return fmt.Errorf("user is already registed")
	}

	store.users[u.ID] = u

	return nil
}

func (store *InMemoryUserStorage) Unregister(ctx context.Context, u *models.User) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.users[u.ID]; !ok {
		return fmt.Errorf("user is not registed")
	}

	delete(store.users, u.ID)
	return nil
}

func NewInMemoryUserStorage() *InMemoryUserStorage {
	store := InMemoryUserStorage{
		mu:    sync.RWMutex{},
		users: make(map[int64]*models.User),
	}
	return &store
}
