package client

import (
	"fmt"
	"strings"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

// FormatMessage da formato legible a un mensaje.
func FormatMessage(m models.Message) string {
	hora := m.SentAt.Local().Format("15:04:05")
	if m.IsDirect() {
		return fmt.Sprintf("[%s] (privado) %s -> %s: %s", hora, m.From, m.To, m.Text)
	}
	return fmt.Sprintf("[%s] [%s] %s: %s", hora, m.Room, m.From, m.Text)
}

// FormatResponse convierte lo que envía el servidor en texto para el usuario.
func FormatResponse(r protocol.Response) string {
	if r.Type == protocol.TypeMessage && r.Message != nil {
		return FormatMessage(*r.Message)
	}
	if !r.OK {
		return "ERROR: " + r.Error
	}
	switch {
	case r.Users != nil:
		return "Conectados: " + strings.Join(r.Users, ", ")
	case r.Rooms != nil:
		parts := make([]string, 0, len(r.Rooms))
		for _, room := range r.Rooms {
			if room.Member {
				parts = append(parts, room.Name+" (miembro)")
			} else {
				parts = append(parts, room.Name)
			}
		}
		return "Salas: " + strings.Join(parts, ", ")
	case r.Cmd == protocol.CmdHistory || r.Cmd == protocol.CmdDMHistory:
		if len(r.Messages) == 0 {
			return "Aún no hay mensajes"
		}
		lines := []string{"Historial (" + r.Info + "):"}
		for _, m := range r.Messages {
			lines = append(lines, "  "+FormatMessage(m))
		}
		return strings.Join(lines, "\n")
	case r.Info != "":
		return "OK: " + r.Info
	default:
		return "OK"
	}
}
