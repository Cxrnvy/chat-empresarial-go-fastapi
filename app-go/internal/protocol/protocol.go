// Package protocol define los mensajes que se intercambian el cliente y el servidor.
// Cada mensaje es un objeto JSON en una sola línea (JSON Lines) sobre una conexión TCP.
package protocol

import "github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"

// Comandos que el servidor entiende.
const (
	CmdLogin     = "login"
	CmdLogout    = "logout"
	CmdQuit      = "quit"
	CmdRooms     = "rooms"
	CmdWho       = "who"
	CmdJoin      = "join"
	CmdSay       = "say"
	CmdDM        = "dm"
	CmdHistory   = "history"
	CmdDMHistory = "dmhistory"
)

// Tipos de respuesta del servidor.
const (
	TypeResponse = "response" // contestación a una solicitud del cliente
	TypeMessage  = "message"  // mensaje entregado en tiempo real, sin que el cliente lo pida
)

// Request es lo que envía el cliente.
type Request struct {
	Cmd   string `json:"cmd"`
	User  string `json:"user,omitempty"`
	Room  string `json:"room,omitempty"`
	To    string `json:"to,omitempty"`
	With  string `json:"with,omitempty"`
	Text  string `json:"text,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// Response es lo que devuelve el servidor, ya sea una respuesta o un mensaje en tiempo real.
type Response struct {
	Type     string            `json:"type"`
	OK       bool              `json:"ok"`
	Cmd      string            `json:"cmd,omitempty"`
	Error    string            `json:"error,omitempty"`
	Info     string            `json:"info,omitempty"`
	Users    []string          `json:"users,omitempty"`
	Rooms    []models.RoomInfo `json:"rooms,omitempty"`
	Messages []models.Message  `json:"messages,omitempty"`
	Message  *models.Message   `json:"message,omitempty"`
}
