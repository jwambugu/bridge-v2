package repository

type Store struct {
	UserRepo     User
	CategoryRepo Category
}

type scanner interface {
	Scan(dest ...any) error
	Err() error
}

func NewStore() Store {
	return Store{}
}
