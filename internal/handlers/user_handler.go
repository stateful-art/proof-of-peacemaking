package handlers

import "proofofpeacemaking/internal/core/ports"

type UserHandler struct {
	userService ports.UserService
}

func NewUserHandler(userService ports.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUserService() ports.UserService {
	return h.userService
}
