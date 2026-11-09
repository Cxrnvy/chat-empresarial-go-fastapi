package services

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/storage"
)

const (
	defaultHistory = 20
	maxHistory     = 100
)

// ChatService reúne las reglas del chat: sesiones, salas, mensajes e historial.
// Se compone de un Store (datos), un Authenticator (identidad) y de las entidades en memoria.
type ChatService struct {
	mu       sync.Mutex
	store    storage.Store
	auth     Authenticator
	users    map[string]*models.User
	rooms    map[string]*models.Room
	sessions map[string]*models.Session
	now      func() time.Time
}

// NewChatService carga usuarios y salas y deja el servicio listo. Si auth es nil usa el directorio
// local, y si now es nil usa la hora del sistema (se puede inyectar un reloj para las pruebas).
func NewChatService(store storage.Store, auth Authenticator, now func() time.Time) (*ChatService, error) {
	users, err := store.LoadUsers()
	if err != nil {
		return nil, fmt.Errorf("no se pudieron cargar los usuarios: %w", err)
	}
	rooms, err := store.LoadRooms()
	if err != nil {
		return nil, fmt.Errorf("no se pudieron cargar las salas: %w", err)
	}
	s := &ChatService{
		store:    store,
		auth:     auth,
		users:    make(map[string]*models.User, len(users)),
		rooms:    make(map[string]*models.Room, len(rooms)),
		sessions: make(map[string]*models.Session),
		now:      now,
	}
	for _, u := range users {
		s.users[u.Username()] = u
	}
	for _, r := range rooms {
		for _, m := range r.Members() {
			if _, ok := s.users[m]; !ok {
				return nil, fmt.Errorf("la sala %q tiene un miembro que no existe: %q", r.Name(), m)
			}
		}
		s.rooms[r.Name()] = r
	}
	if s.auth == nil {
		s.auth = NewDirectoryAuthenticator(users)
	}
	if s.now == nil {
		s.now = time.Now
	}
	return s, nil
}

// Login abre una sesión para el usuario. La credencial se pasa tal cual al Authenticator.
func (s *ChatService) Login(username, credential, remoteAddr string) (*models.Session, error) {
	user, err := s.auth.Authenticate(username, credential)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[username]; ok {
		return nil, models.ErrAlreadyLoggedIn
	}
	session, err := models.NewSession(user, remoteAddr, s.now())
	if err != nil {
		return nil, err
	}
	s.sessions[username] = session
	return session, nil
}

// Logout cierra la sesión del usuario.
func (s *ChatService) Logout(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[username]; !ok {
		return models.ErrNotLoggedIn
	}
	delete(s.sessions, username)
	return nil
}

// OnlineUsers devuelve, en orden alfabético, quién tiene sesión abierta.
func (s *ChatService) OnlineUsers() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.sessions))
	for name := range s.sessions {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ListRooms devuelve todas las salas e indica en cuáles es miembro el usuario.
func (s *ChatService) ListRooms(username string) ([]models.RoomInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(username); err != nil {
		return nil, err
	}
	out := make([]models.RoomInfo, 0, len(s.rooms))
	for name, r := range s.rooms {
		out = append(out, models.RoomInfo{Name: name, Member: r.IsMember(username)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// JoinRoom agrega al usuario a una sala existente. En el prototipo el cambio vive solo en memoria.
func (s *ChatService) JoinRoom(username, room string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(username); err != nil {
		return err
	}
	r, ok := s.rooms[room]
	if !ok {
		return models.ErrUnknownRoom
	}
	return r.AddMember(username)
}

// SendToRoom valida y guarda un mensaje de sala. Devuelve el mensaje guardado y los miembros de la
// sala, para que quien llame sepa a quién entregarlo.
func (s *ChatService) SendToRoom(from, room, text string) (models.Message, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(from); err != nil {
		return models.Message{}, nil, err
	}
	r, ok := s.rooms[room]
	if !ok {
		return models.Message{}, nil, models.ErrUnknownRoom
	}
	if !r.IsMember(from) {
		return models.Message{}, nil, models.ErrNotMember
	}
	msg, err := models.NewRoomMessage(s.store.NextMessageID(), from, room, text, s.now())
	if err != nil {
		return models.Message{}, nil, err
	}
	if err := s.store.AppendMessage(msg); err != nil {
		return models.Message{}, nil, err
	}
	return msg, r.Members(), nil
}

// SendDirect valida y guarda un mensaje directo.
func (s *ChatService) SendDirect(from, to, text string) (models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(from); err != nil {
		return models.Message{}, err
	}
	if !models.ValidUsername(to) {
		return models.Message{}, models.ErrInvalidUsername
	}
	if _, ok := s.users[to]; !ok {
		return models.Message{}, models.ErrUnknownUser
	}
	if to == from {
		return models.Message{}, models.ErrSelfMessage
	}
	msg, err := models.NewDirectMessage(s.store.NextMessageID(), from, to, text, s.now())
	if err != nil {
		return models.Message{}, err
	}
	if err := s.store.AppendMessage(msg); err != nil {
		return models.Message{}, err
	}
	return msg, nil
}

// RoomHistory devuelve los últimos mensajes de una sala, de los más antiguos a los más recientes.
func (s *ChatService) RoomHistory(requester, room string, limit int) ([]models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(requester); err != nil {
		return nil, err
	}
	r, ok := s.rooms[room]
	if !ok {
		return nil, models.ErrUnknownRoom
	}
	if !r.IsMember(requester) {
		return nil, models.ErrNotMember
	}
	all, err := s.store.Messages()
	if err != nil {
		return nil, err
	}
	var out []models.Message
	for _, m := range all {
		if m.Room == room {
			out = append(out, m)
		}
	}
	return lastN(out, limit), nil
}

// DirectHistory devuelve los últimos mensajes directos entre el solicitante y otra persona.
func (s *ChatService) DirectHistory(requester, other string, limit int) ([]models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireSession(requester); err != nil {
		return nil, err
	}
	if !models.ValidUsername(other) {
		return nil, models.ErrInvalidUsername
	}
	if _, ok := s.users[other]; !ok {
		return nil, models.ErrUnknownUser
	}
	all, err := s.store.Messages()
	if err != nil {
		return nil, err
	}
	var out []models.Message
	for _, m := range all {
		if m.BetweenUsers(requester, other) {
			out = append(out, m)
		}
	}
	return lastN(out, limit), nil
}

// requireSession debe llamarse con el candado tomado.
func (s *ChatService) requireSession(username string) error {
	if _, ok := s.sessions[username]; !ok {
		return models.ErrNotLoggedIn
	}
	return nil
}

// lastN se queda con los últimos n elementos, aplicando un valor por defecto y un tope.
func lastN(msgs []models.Message, n int) []models.Message {
	if n <= 0 {
		n = defaultHistory
	}
	if n > maxHistory {
		n = maxHistory
	}
	if len(msgs) > n {
		return msgs[len(msgs)-n:]
	}
	return msgs
}
