package services

import (
	"user-service/repositories"
	services "user-service/services/user"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
}

type IServicesRegistry interface {
	GetUser() services.IUserServices
}

func NewServicesRegistry(repository repositories.IRepositoryRegistry) IServicesRegistry {
	return &Registry{repository: repository}
}
func (r *Registry) GetUser() services.IUserServices {
	return services.NewUserService(r.repository)
}
