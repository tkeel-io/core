package repository

type IRepository interface {
}

type Repository struct{}

func New() *Repository {
	return &Repository{}
}
