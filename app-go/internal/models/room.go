package models

import (
	"fmt"
	"regexp"
	"sort"
)

var roomPattern = regexp.MustCompile(`^[a-z0-9_-]{3,30}$`)

// ValidRoomName indica si el nombre de sala cumple el formato permitido.
func ValidRoomName(name string) bool {
	return roomPattern.MatchString(name)
}

// RoomInfo es el resumen de una sala desde el punto de vista de un usuario.
type RoomInfo struct {
	Name   string `json:"name"`
	Member bool   `json:"member"`
}

// Room es una sala de conversación con su lista de miembros.
type Room struct {
	name    string
	members map[string]struct{}
}

// NewRoom crea una sala válida con sus miembros iniciales.
func NewRoom(name string, members []string) (*Room, error) {
	if !ValidRoomName(name) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRoom, name)
	}
	r := &Room{name: name, members: make(map[string]struct{})}
	for _, m := range members {
		if err := r.AddMember(m); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Room) Name() string { return r.name }

// AddMember agrega a un usuario a la sala. Agregar dos veces al mismo usuario no es un error.
func (r *Room) AddMember(username string) error {
	if !ValidUsername(username) {
		return fmt.Errorf("%w: %q", ErrInvalidUsername, username)
	}
	r.members[username] = struct{}{}
	return nil
}

func (r *Room) IsMember(username string) bool {
	_, ok := r.members[username]
	return ok
}

// Members devuelve una copia ordenada de los miembros, para que nadie modifique el mapa interno.
func (r *Room) Members() []string {
	out := make([]string, 0, len(r.members))
	for m := range r.members {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}
