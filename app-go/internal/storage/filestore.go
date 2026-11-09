package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
)

const (
	usersFile    = "users.json"
	roomsFile    = "rooms.json"
	messagesFile = "messages.jsonl"
)

// userRecord y roomRecord son la forma en que se guardan los datos en disco.
// Se separan de models.User y models.Room para que el modelo pueda tener campos privados.
type userRecord struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

type roomRecord struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// FileStore guarda usuarios y salas en JSON (solo lectura) y los mensajes en un archivo
// messages.jsonl, con un mensaje por línea. Es seguro para varias goroutines.
type FileStore struct {
	dir    string
	mu     sync.Mutex
	lastID int
}

// NewFileStore abre un directorio de datos y recupera el último identificador de mensaje usado.
func NewFileStore(dir string) (*FileStore, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("el directorio de datos no es accesible (%s): %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s no es un directorio", dir)
	}
	fs := &FileStore{dir: dir}
	msgs, err := fs.readMessages()
	if err != nil {
		return nil, err
	}
	for _, m := range msgs {
		if m.ID > fs.lastID {
			fs.lastID = m.ID
		}
	}
	return fs, nil
}

func (f *FileStore) LoadUsers() ([]*models.User, error) {
	var records []userRecord
	if err := f.readJSON(usersFile, &records); err != nil {
		return nil, err
	}
	users := make([]*models.User, 0, len(records))
	for i, r := range records {
		u, err := models.NewUser(r.Username, models.Role(r.Role), r.Active)
		if err != nil {
			return nil, fmt.Errorf("%s, registro %d: %w", usersFile, i+1, err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (f *FileStore) LoadRooms() ([]*models.Room, error) {
	var records []roomRecord
	if err := f.readJSON(roomsFile, &records); err != nil {
		return nil, err
	}
	rooms := make([]*models.Room, 0, len(records))
	for i, r := range records {
		room, err := models.NewRoom(r.Name, r.Members)
		if err != nil {
			return nil, fmt.Errorf("%s, registro %d: %w", roomsFile, i+1, err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (f *FileStore) NextMessageID() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastID + 1
}

// AppendMessage agrega un mensaje al final del archivo. El archivo se crea con permisos 0600
// porque contiene conversaciones privadas.
func (f *FileStore) AppendMessage(msg models.Message) error {
	line, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("no se pudo convertir el mensaje a JSON: %w", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	file, err := os.OpenFile(filepath.Join(f.dir, messagesFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo de mensajes: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("no se pudo escribir el mensaje: %w", err)
	}
	if msg.ID > f.lastID {
		f.lastID = msg.ID
	}
	return nil
}

func (f *FileStore) Messages() ([]models.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.readMessages()
}

func (f *FileStore) readJSON(name string, target any) error {
	data, err := os.ReadFile(filepath.Join(f.dir, name))
	if err != nil {
		return fmt.Errorf("no se pudo leer %s: %w", name, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s tiene un formato inválido: %w", name, err)
	}
	return nil
}

// readMessages lee todo el historial. El llamador debe tener el candado (o estar en el constructor).
func (f *FileStore) readMessages() ([]models.Message, error) {
	file, err := os.Open(filepath.Join(f.dir, messagesFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // todavía no hay mensajes
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo de mensajes: %w", err)
	}
	defer file.Close()

	var msgs []models.Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		var m models.Message
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("%s, línea %d: formato inválido: %w", messagesFile, line, err)
		}
		msgs = append(msgs, m)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error al leer el archivo de mensajes: %w", err)
	}
	return msgs, nil
}
