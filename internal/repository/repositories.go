package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	Ad *AdRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Ad: NewAdRepository(db),
	}
}
