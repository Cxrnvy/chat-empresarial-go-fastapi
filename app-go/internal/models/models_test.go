package models

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	cases := []struct {
		name    string
		user    string
		role    Role
		wantErr error
	}{
		{"usuario válido", "ariel", RoleUser, nil},
		{"administrador válido", "cesar", RoleAdmin, nil},
		{"nombre muy corto", "ab", RoleUser, ErrInvalidUsername},
		{"mayúsculas no permitidas", "Ariel", RoleUser, ErrInvalidUsername},
		{"símbolos no permitidos", "ari el", RoleUser, ErrInvalidUsername},
		{"rol inexistente", "ariel", Role("root"), ErrInvalidRole},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := NewUser(tc.user, tc.role, true)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, se esperaba %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && (u.Username() != tc.user || !u.Active()) {
				t.Fatalf("el usuario no quedó con el estado esperado: %+v", u)
			}
		})
	}
}

func TestUserSetActive(t *testing.T) {
	u, _ := NewUser("ariel", RoleUser, true)
	u.SetActive(false)
	if u.Active() {
		t.Fatal("SetActive(false) no desactivó la cuenta")
	}
}

func TestRoomMembers(t *testing.T) {
	r, err := NewRoom("general", []string{"jorge", "ariel"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsMember("ariel") || r.IsMember("cesar") {
		t.Fatal("IsMember devolvió un resultado incorrecto")
	}
	if got := strings.Join(r.Members(), ","); got != "ariel,jorge" {
		t.Fatalf("Members() = %q, se esperaba orden alfabético", got)
	}
	// Modificar la copia no debe afectar a la sala.
	r.Members()[0] = "intruso"
	if r.IsMember("intruso") {
		t.Fatal("Members() expone el estado interno de la sala")
	}
	if err := r.AddMember("Mal Nombre"); !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("AddMember aceptó un nombre inválido: %v", err)
	}
	if _, err := NewRoom("x", nil); !errors.Is(err, ErrInvalidRoom) {
		t.Fatalf("NewRoom aceptó un nombre inválido: %v", err)
	}
}

func TestValidateText(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		want    string
		wantErr error
	}{
		{"texto normal", "hola equipo", "hola equipo", nil},
		{"se recortan espacios", "  hola  ", "hola", nil},
		{"vacío", "", "", ErrEmptyMessage},
		{"solo espacios", "   \t ", "", ErrEmptyMessage},
		{"exactamente el máximo", strings.Repeat("a", MaxMessageLength), strings.Repeat("a", MaxMessageLength), nil},
		{"un carácter de más", strings.Repeat("a", MaxMessageLength+1), "", ErrMessageTooLong},
		{"cuenta caracteres y no bytes", strings.Repeat("ñ", MaxMessageLength), strings.Repeat("ñ", MaxMessageLength), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateText(tc.text)
			if !errors.Is(err, tc.wantErr) || got != tc.want {
				t.Fatalf("ValidateText() = (%q, %v), se esperaba (%q, %v)", got, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestMessageConstructors(t *testing.T) {
	at := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	m, err := NewRoomMessage(1, "ariel", "general", " hola ", at)
	if err != nil || m.IsDirect() || m.Text != "hola" {
		t.Fatalf("mensaje de sala incorrecto: %+v, %v", m, err)
	}
	d, err := NewDirectMessage(2, "ariel", "jorge", "hola", at)
	if err != nil || !d.IsDirect() {
		t.Fatalf("mensaje directo incorrecto: %+v, %v", d, err)
	}
	if !d.BetweenUsers("jorge", "ariel") || d.BetweenUsers("ariel", "cesar") {
		t.Fatal("BetweenUsers devolvió un resultado incorrecto")
	}
	if _, err := NewRoomMessage(3, "ariel", "general", "  ", at); !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("se aceptó un mensaje vacío: %v", err)
	}
}

func TestNewSessionRequiresUser(t *testing.T) {
	if _, err := NewSession(nil, "127.0.0.1:1", time.Now()); err == nil {
		t.Fatal("NewSession aceptó un usuario nil")
	}
	u, _ := NewUser("ariel", RoleUser, true)
	s, err := NewSession(u, "127.0.0.1:1", time.Now())
	if err != nil || s.Username() != "ariel" {
		t.Fatalf("sesión incorrecta: %v", err)
	}
}
