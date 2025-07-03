package storage

type Store struct {
	Users UserStorage
	Tasks Taskstorage
}

func NewStore() *Store {
	return &Store{
		Users: NewInMemoryUserStorage(),
		Tasks: NewInMemoryTaskStorage(),
	}
}
