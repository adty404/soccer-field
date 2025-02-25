package seeder

import "gorm.io/gorm"

type Registry struct {
	db *gorm.DB
}

type ISeederRegistry interface {
	Run()
}

func NewSeederRegistry(db *gorm.DB) ISeederRegistry {
	return &Registry{db: db}
}

func (S *Registry) Run() {
	RunRoleSeeder(S.db)
	RunUserSeeder(S.db)
}
