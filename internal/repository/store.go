package repository

type Store struct {
	UserRepo     User
	CategoryRepo Category
}

func NewStore() Store {
	return Store{}
}
