package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/storage"
)

func newService(t *testing.T) *ChatService {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"users.json": `[{"username":"cesar","role":"admin","active":true},
			{"username":"ariel","role":"user","active":true},
			{"username":"jorge","role":"user","active":true},
			{"username":"lucia","role":"user","active":false}]`,
		"rooms.json": `[{"name":"general","members":["cesar","ariel","jorge"]},
			{"name":"proyectos","members":["cesar","ariel"]}]`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	store, err := storage.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	svc, err := NewChatService(store, nil, func() time.Time { return fixed })
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func login(t *testing.T, svc *ChatService, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, err := svc.Login(n, "", "127.0.0.1:1"); err != nil {
			t.Fatalf("login de %s: %v", n, err)
		}
	}
}

func TestLoginLogout(t *testing.T) {
	svc := newService(t)
	cases := []struct {
		user    string
		wantErr error
	}{
		{"ariel", nil},
		{"fantasma", models.ErrUnknownUser},
		{"lucia", models.ErrInactiveUser},
		{"Mal Nombre", models.ErrInvalidUsername},
	}
	for _, tc := range cases {
		if _, err := svc.Login(tc.user, "", "127.0.0.1:1"); !errors.Is(err, tc.wantErr) {
			t.Fatalf("login(%q) = %v, se esperaba %v", tc.user, err, tc.wantErr)
		}
	}
	if _, err := svc.Login("ariel", "", "127.0.0.1:2"); !errors.Is(err, models.ErrAlreadyLoggedIn) {
		t.Fatalf("un segundo login del mismo usuario debía rechazarse: %v", err)
	}
	if got := svc.OnlineUsers(); len(got) != 1 || got[0] != "ariel" {
		t.Fatalf("OnlineUsers() = %v", got)
	}
	if err := svc.Logout("ariel"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout("ariel"); !errors.Is(err, models.ErrNotLoggedIn) {
		t.Fatalf("logout sin sesión debía fallar: %v", err)
	}
}

func TestSendToRoom(t *testing.T) {
	svc := newService(t)
	login(t, svc, "ariel", "jorge")

	msg, members, err := svc.SendToRoom("ariel", "general", "  hola equipo  ")
	if err != nil || msg.Text != "hola equipo" || msg.ID != 1 || len(members) != 3 {
		t.Fatalf("envío correcto falló: %+v, %v, %v", msg, members, err)
	}
	cases := []struct {
		name    string
		from    string
		room    string
		text    string
		wantErr error
	}{
		{"mensaje vacío", "ariel", "general", "   ", models.ErrEmptyMessage},
		{"sala inexistente", "ariel", "nada", "hola", models.ErrUnknownRoom},
		{"no es miembro", "jorge", "proyectos", "hola", models.ErrNotMember},
		{"sin sesión", "cesar", "general", "hola", models.ErrNotLoggedIn},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := svc.SendToRoom(tc.from, tc.room, tc.text); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, se esperaba %v", err, tc.wantErr)
			}
		})
	}
}

func TestJoinRoom(t *testing.T) {
	svc := newService(t)
	login(t, svc, "jorge")
	if _, _, err := svc.SendToRoom("jorge", "proyectos", "hola"); !errors.Is(err, models.ErrNotMember) {
		t.Fatalf("antes de unirse debía rechazarse: %v", err)
	}
	if err := svc.JoinRoom("jorge", "proyectos"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.SendToRoom("jorge", "proyectos", "hola"); err != nil {
		t.Fatalf("después de unirse debía permitirse: %v", err)
	}
	if err := svc.JoinRoom("jorge", "nada"); !errors.Is(err, models.ErrUnknownRoom) {
		t.Fatalf("unirse a una sala inexistente debía fallar: %v", err)
	}
}

func TestSendDirect(t *testing.T) {
	svc := newService(t)
	login(t, svc, "ariel")
	if _, err := svc.SendDirect("ariel", "jorge", "hola"); err != nil {
		t.Fatalf("mensaje directo válido falló: %v", err)
	}
	cases := []struct {
		to      string
		text    string
		wantErr error
	}{
		{"fantasma", "hola", models.ErrUnknownUser},
		{"ariel", "hola", models.ErrSelfMessage},
		{"jorge", "", models.ErrEmptyMessage},
		{"Mal Nombre", "hola", models.ErrInvalidUsername},
	}
	for _, tc := range cases {
		if _, err := svc.SendDirect("ariel", tc.to, tc.text); !errors.Is(err, tc.wantErr) {
			t.Fatalf("SendDirect(%q, %q) = %v, se esperaba %v", tc.to, tc.text, err, tc.wantErr)
		}
	}
}

func TestHistory(t *testing.T) {
	svc := newService(t)
	login(t, svc, "ariel", "jorge")
	for _, text := range []string{"uno", "dos", "tres"} {
		if _, _, err := svc.SendToRoom("ariel", "general", text); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := svc.SendToRoom("ariel", "proyectos", "solo proyectos"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SendDirect("ariel", "jorge", "privado"); err != nil {
		t.Fatal(err)
	}

	msgs, err := svc.RoomHistory("jorge", "general", 2)
	if err != nil || len(msgs) != 2 || msgs[0].Text != "dos" || msgs[1].Text != "tres" {
		t.Fatalf("historial de sala incorrecto: %+v, %v", msgs, err)
	}
	if _, err := svc.RoomHistory("jorge", "proyectos", 10); !errors.Is(err, models.ErrNotMember) {
		t.Fatalf("un no miembro no debe ver el historial: %v", err)
	}
	dm, err := svc.DirectHistory("jorge", "ariel", 10)
	if err != nil || len(dm) != 1 || dm[0].Text != "privado" {
		t.Fatalf("historial directo incorrecto: %+v, %v", dm, err)
	}
	if _, err := svc.DirectHistory("jorge", "fantasma", 10); !errors.Is(err, models.ErrUnknownUser) {
		t.Fatalf("historial con usuario inexistente debía fallar: %v", err)
	}
	// Una tercera persona no ve la conversación privada.
	login(t, svc, "cesar")
	other, err := svc.DirectHistory("cesar", "ariel", 10)
	if err != nil || len(other) != 0 {
		t.Fatalf("cesar no debía ver mensajes privados ajenos: %+v, %v", other, err)
	}
}

func TestListRooms(t *testing.T) {
	svc := newService(t)
	if _, err := svc.ListRooms("ariel"); !errors.Is(err, models.ErrNotLoggedIn) {
		t.Fatalf("listar salas sin sesión debía fallar: %v", err)
	}
	login(t, svc, "jorge")
	rooms, err := svc.ListRooms("jorge")
	if err != nil || len(rooms) != 2 || !rooms[0].Member || rooms[1].Member {
		t.Fatalf("lista de salas incorrecta: %+v, %v", rooms, err)
	}
}
