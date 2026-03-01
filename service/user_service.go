package service

import "github.com/your-name/roadmap/api/adapter"

type UserService struct {
	userAdapter *adapter.UserAdapter
}

func NewUserService(userAdapter *adapter.UserAdapter) *UserService {
	return &UserService{userAdapter: userAdapter}
}
