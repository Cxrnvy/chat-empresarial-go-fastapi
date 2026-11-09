package models

import (
	"errors"
	"time"
)

// Session representa una sesión activa. Se compone de un *User en lugar de copiar sus datos,
// así la sesión siempre refleja el estado actual de la cuenta.
type Session struct {
	user       *User
	remoteAddr string
	startedAt  time.Time
}

// NewSession crea una sesión para un usuario existente.
func NewSession(user *User, remoteAddr string, startedAt time.Time) (*Session, error) {
	if user == nil {
		return nil, errors.New("no se puede crear una sesión sin usuario")
	}
	return &Session{user: user, remoteAddr: remoteAddr, startedAt: startedAt}, nil
}

func (s *Session) User() *User          { return s.user }
func (s *Session) Username() string     { return s.user.Username() }
func (s *Session) RemoteAddr() string   { return s.remoteAddr }
func (s *Session) StartedAt() time.Time { return s.startedAt }
