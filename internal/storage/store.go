package storage

type Store struct {
	Users UserStorage
}

func NewStore() *Store {
	return &Store{
		Users: NewInMemoryUserStorage(),
	}
}
