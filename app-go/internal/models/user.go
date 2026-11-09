package models

import (
	"fmt"
	"regexp"
)

// Role identifica el nivel de permisos de una cuenta.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

// ValidUsername indica si el nombre cumple el formato permitido.
func ValidUsername(name string) bool {
	return usernamePattern.MatchString(name)
}

// User representa a una persona con cuenta en el chat.
// Sus campos no son exportados: solo se leen con métodos y solo se cambian con SetActive,
// para que ninguna otra capa pueda dejar al usuario en un estado inválido.
type User struct {
	username string
	role     Role
	active   bool
}

// NewUser crea un usuario con un estado válido o devuelve un error.
func NewUser(username string, role Role, active bool) (*User, error) {
	if !ValidUsername(username) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidUsername, username)
	}
	if role != RoleUser && role != RoleAdmin {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRole, string(role))
	}
	return &User{username: username, role: role, active: active}, nil
}

func (u *User) Username() string { return u.username }
func (u *User) Role() Role       { return u.role }
func (u *User) Active() bool     { return u.active }
func (u *User) IsAdmin() bool    { return u.role == RoleAdmin }

// SetActive activa o desactiva la cuenta. Usa receptor por puntero porque modifica el estado.
func (u *User) SetActive(active bool) { u.active = active }
