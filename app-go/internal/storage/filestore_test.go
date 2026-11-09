package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
)

func seedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	users := `[{"username":"ariel","role":"user","active":true},{"username":"cesar","role":"admin","active":true}]`
	rooms := `[{"name":"general","members":["ariel","cesar"]}]`
	if err := os.WriteFile(filepath.Join(dir, "users.json"), []byte(users), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rooms.json"), []byte(rooms), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadUsersAndRooms(t *testing.T) {
	fs, err := NewFileStore(seedDir(t))
	if err != nil {
		t.Fatal(err)
	}
	users, err := fs.LoadUsers()
	if err != nil || len(users) != 2 || !users[1].IsAdmin() {
		t.Fatalf("usuarios cargados incorrectamente: %v, %v", users, err)
	}
	rooms, err := fs.LoadRooms()
	if err != nil || len(rooms) != 1 || !rooms[0].IsMember("cesar") {
		t.Fatalf("salas cargadas incorrectamente: %v, %v", rooms, err)
	}
}

func TestInvalidDataIsRejected(t *testing.T) {
	dir := seedDir(t)
	bad := `[{"username":"Mal Nombre","role":"user","active":true}]`
	if err := os.WriteFile(filepath.Join(dir, "users.json"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	fs, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.LoadUsers(); !errors.Is(err, models.ErrInvalidUsername) {
		t.Fatalf("se esperaba ErrInvalidUsername y se obtuvo %v", err)
	}
}

func TestMissingDirectory(t *testing.T) {
	if _, err := NewFileStore(filepath.Join(t.TempDir(), "no-existe")); err == nil {
		t.Fatal("se aceptó un directorio inexistente")
	}
}

func TestAppendAndReadMessages(t *testing.T) {
	dir := seedDir(t)
	fs, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := fs.NextMessageID(); got != 1 {
		t.Fatalf("el primer id debía ser 1 y fue %d", got)
	}
	at := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	for i := 1; i <= 3; i++ {
		m, _ := models.NewRoomMessage(fs.NextMessageID(), "ariel", "general", "hola", at)
		if err := fs.AppendMessage(m); err != nil {
			t.Fatal(err)
		}
	}
	msgs, err := fs.Messages()
	if err != nil || len(msgs) != 3 || msgs[2].ID != 3 {
		t.Fatalf("mensajes leídos incorrectamente: %v, %v", msgs, err)
	}
	// Al reabrir el almacenamiento, los ids siguen donde quedaron.
	reopened, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.NextMessageID(); got != 4 {
		t.Fatalf("tras reabrir, el siguiente id debía ser 4 y fue %d", got)
	}
}

func TestCorruptMessagesFile(t *testing.T) {
	dir := seedDir(t)
	if err := os.WriteFile(filepath.Join(dir, "messages.jsonl"), []byte("esto no es json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileStore(dir); err == nil {
		t.Fatal("se aceptó un archivo de mensajes corrupto")
	}
}
