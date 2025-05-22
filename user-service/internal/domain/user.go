package domain

import (
	"time"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserFilter struct {
	Email    string
	Username string
	Role     UserRole
	Page     int
	PageSize int
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

func (u *User) IsUser() bool {
	return u.Role == UserRoleUser
}
