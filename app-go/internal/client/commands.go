package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

// Action dice qué hay que hacer con una línea escrita por el usuario.
type Action int

const (
	ActionNone    Action = iota // línea vacía
	ActionSend                  // enviar una solicitud al servidor
	ActionSetRoom               // cambiar la sala actual (solo local)
	ActionHelp                  // mostrar la ayuda
	ActionQuit                  // salir
)

// Command es el resultado de interpretar una línea.
type Command struct {
	Action  Action
	Request protocol.Request
	Room    string
}

// HelpText es la ayuda que muestra el cliente.
const HelpText = `Comandos:
  /login <usuario>           inicia sesión
  /logout                    cierra la sesión
  /rooms                     lista las salas
  /join <sala>               te une a una sala
  /room <sala>               elige la sala a la que van tus mensajes
  /say <sala> <texto>        envía un mensaje a una sala
  /dm <usuario> <texto>      envía un mensaje directo
  /history [sala] [n]        últimos n mensajes de una sala
  /dmhistory <usuario> [n]   últimos n mensajes directos con esa persona
  /who                       usuarios conectados
  /help                      muestra esta ayuda
  /quit                      sale del programa
Si escribes texto sin comando, se envía a la sala elegida con /room.`

// ParseCommand interpreta una línea del usuario. No valida el contenido de los mensajes:
// eso lo hace el servidor, que es quien debe desconfiar de lo que recibe.
func ParseCommand(line, currentRoom string) (Command, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{Action: ActionNone}, nil
	}
	if !strings.HasPrefix(line, "/") {
		if currentRoom == "" {
			return Command{}, errors.New("elige una sala con /room <sala> o usa /say <sala> <texto>")
		}
		return send(protocol.Request{Cmd: protocol.CmdSay, Room: currentRoom, Text: line}), nil
	}
	name, rest := splitFirst(line[1:])
	switch name {
	case "help":
		return Command{Action: ActionHelp}, nil
	case "quit", "exit":
		return Command{Action: ActionQuit}, nil
	case "login":
		if rest == "" || strings.Contains(rest, " ") {
			return Command{}, usage("/login <usuario>")
		}
		return send(protocol.Request{Cmd: protocol.CmdLogin, User: rest}), nil
	case "logout":
		return send(protocol.Request{Cmd: protocol.CmdLogout}), nil
	case "rooms":
		return send(protocol.Request{Cmd: protocol.CmdRooms}), nil
	case "who":
		return send(protocol.Request{Cmd: protocol.CmdWho}), nil
	case "join":
		if rest == "" || strings.Contains(rest, " ") {
			return Command{}, usage("/join <sala>")
		}
		return send(protocol.Request{Cmd: protocol.CmdJoin, Room: rest}), nil
	case "room":
		if rest == "" || strings.Contains(rest, " ") {
			return Command{}, usage("/room <sala>")
		}
		return Command{Action: ActionSetRoom, Room: rest}, nil
	case "say":
		room, text := splitFirst(rest)
		if room == "" {
			return Command{}, usage("/say <sala> <texto>")
		}
		return send(protocol.Request{Cmd: protocol.CmdSay, Room: room, Text: text}), nil
	case "dm":
		to, text := splitFirst(rest)
		if to == "" {
			return Command{}, usage("/dm <usuario> <texto>")
		}
		return send(protocol.Request{Cmd: protocol.CmdDM, To: to, Text: text}), nil
	case "history":
		room, limit := currentRoom, 0
		for _, f := range strings.Fields(rest) {
			if n, err := strconv.Atoi(f); err == nil {
				limit = n
			} else {
				room = f
			}
		}
		if room == "" {
			return Command{}, usage("/history <sala> [n]")
		}
		return send(protocol.Request{Cmd: protocol.CmdHistory, Room: room, Limit: limit}), nil
	case "dmhistory":
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			return Command{}, usage("/dmhistory <usuario> [n]")
		}
		limit := 0
		if len(fields) > 1 {
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				return Command{}, usage("/dmhistory <usuario> [n]")
			}
			limit = n
		}
		return send(protocol.Request{Cmd: protocol.CmdDMHistory, With: fields[0], Limit: limit}), nil
	default:
		return Command{}, fmt.Errorf("comando desconocido: /%s (escribe /help para ver los comandos)", name)
	}
}

func send(req protocol.Request) Command { return Command{Action: ActionSend, Request: req} }

func usage(u string) error { return fmt.Errorf("uso: %s", u) }

// splitFirst separa la primera palabra del resto del texto.
func splitFirst(s string) (string, string) {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, ' ')
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i+1:])
}
