package controllers

import (
	controllers "user-service/controllers/user"
	"user-service/services"
)

type Registry struct {
	service services.IServicesRegistry
}

type IControllerRegistry interface {
	GetUserController() controllers.IUserController
}

func NewControllerRegistry(service services.IServicesRegistry) IControllerRegistry {
	return &Registry{service: service}
}

func (r *Registry) GetUserController() controllers.IUserController {
	return controllers.NewUserController(r.service)
}
