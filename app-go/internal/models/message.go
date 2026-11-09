package models

import (
	"strings"
	"time"
	"unicode/utf8"
)

// MaxMessageLength es el largo máximo de un mensaje, en caracteres.
const MaxMessageLength = 1000

// Message es un mensaje de sala (campo Room) o directo (campo To).
// La hora la asigna siempre el servidor, nunca el cliente.
type Message struct {
	ID     int       `json:"id"`
	From   string    `json:"from"`
	Room   string    `json:"room,omitempty"`
	To     string    `json:"to,omitempty"`
	Text   string    `json:"text"`
	SentAt time.Time `json:"sent_at"`
}

// ValidateText limpia los espacios sobrantes y comprueba que el texto no esté vacío ni sea muy largo.
func ValidateText(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", ErrEmptyMessage
	}
	if utf8.RuneCountInString(trimmed) > MaxMessageLength {
		return "", ErrMessageTooLong
	}
	return trimmed, nil
}

// NewRoomMessage crea un mensaje de sala válido.
func NewRoomMessage(id int, from, room, text string, at time.Time) (Message, error) {
	clean, err := ValidateText(text)
	if err != nil {
		return Message{}, err
	}
	return Message{ID: id, From: from, Room: room, Text: clean, SentAt: at}, nil
}

// NewDirectMessage crea un mensaje directo válido.
func NewDirectMessage(id int, from, to, text string, at time.Time) (Message, error) {
	clean, err := ValidateText(text)
	if err != nil {
		return Message{}, err
	}
	return Message{ID: id, From: from, To: to, Text: clean, SentAt: at}, nil
}

// IsDirect indica si es un mensaje directo.
func (m Message) IsDirect() bool { return m.To != "" }

// BetweenUsers indica si el mensaje directo es entre esas dos personas, en cualquier sentido.
func (m Message) BetweenUsers(a, b string) bool {
	return m.IsDirect() && ((m.From == a && m.To == b) || (m.From == b && m.To == a))
}
