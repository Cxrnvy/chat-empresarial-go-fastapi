package handlers

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/services"
)

const maxLineBytes = 64 * 1024

// Handler atiende las conexiones de los clientes: lee solicitudes, llama al servicio y responde.
type Handler struct {
	svc    *services.ChatService
	hub    *Hub
	logger *log.Logger
}

// NewHandler crea un Handler. Si logger es nil, usa el logger estándar.
func NewHandler(svc *services.ChatService, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.Default()
	}
	return &Handler{svc: svc, hub: NewHub(), logger: logger}
}

// ServeConn atiende una conexión hasta que el cliente se va. Se ejecuta en su propia goroutine.
func (h *Handler) ServeConn(conn net.Conn) {
	defer conn.Close()
	client := newConnClient(conn)
	remote := conn.RemoteAddr().String()
	h.event("conexion", "-", remote, "abierta", "")

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4096), maxLineBytes)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var req protocol.Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			h.event("solicitud", client.username, remote, "rechazada", "json mal formado")
			if client.Send(errorResponse("", fmt.Errorf("%w: el formato JSON no es válido", models.ErrInvalidRequest))) != nil {
				break
			}
			continue
		}
		resp, quit := h.dispatch(client, remote, req)
		if err := client.Send(resp); err != nil {
			break
		}
		if quit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		h.event("conexion", client.username, remote, "error", err.Error())
	}
	h.release(client, remote)
}

// release cierra la sesión si el cliente se desconectó sin hacer logout.
func (h *Handler) release(c *connClient, remote string) {
	if c.username != "" {
		_ = h.svc.Logout(c.username)
		h.hub.unregister(c.username, c)
		h.event("logout", c.username, remote, "desconexion", "")
		c.username = ""
	}
	h.event("conexion", "-", remote, "cerrada", "")
}

// dispatch ejecuta un comando. El segundo valor indica si hay que cerrar la conexión.
func (h *Handler) dispatch(c *connClient, remote string, req protocol.Request) (protocol.Response, bool) {
	switch req.Cmd {
	case protocol.CmdLogin:
		return h.login(c, remote, req), false
	case protocol.CmdLogout:
		return h.logout(c, remote, req.Cmd), false
	case protocol.CmdQuit:
		resp := h.logout(c, remote, req.Cmd)
		if !resp.OK {
			resp = okResponse(req.Cmd, "hasta pronto")
		}
		return resp, true
	case protocol.CmdWho:
		if c.username == "" {
			return errorResponse(req.Cmd, models.ErrNotLoggedIn), false
		}
		resp := okResponse(req.Cmd, "")
		resp.Users = h.svc.OnlineUsers()
		return resp, false
	case protocol.CmdRooms:
		rooms, err := h.svc.ListRooms(c.username)
		if err != nil {
			return h.fail(c, remote, req.Cmd, err), false
		}
		resp := okResponse(req.Cmd, "")
		resp.Rooms = rooms
		return resp, false
	case protocol.CmdJoin:
		if err := h.svc.JoinRoom(c.username, req.Room); err != nil {
			return h.fail(c, remote, req.Cmd, err), false
		}
		h.event("join", c.username, remote, "ok", "sala="+req.Room)
		return okResponse(req.Cmd, "te uniste a la sala "+req.Room), false
	case protocol.CmdSay:
		return h.say(c, remote, req), false
	case protocol.CmdDM:
		return h.direct(c, remote, req), false
	case protocol.CmdHistory:
		msgs, err := h.svc.RoomHistory(c.username, req.Room, req.Limit)
		if err != nil {
			return h.fail(c, remote, req.Cmd, err), false
		}
		resp := okResponse(req.Cmd, countMessages(len(msgs)))
		resp.Messages = msgs
		h.event("historial", c.username, remote, "ok", "sala="+req.Room)
		return resp, false
	case protocol.CmdDMHistory:
		msgs, err := h.svc.DirectHistory(c.username, req.With, req.Limit)
		if err != nil {
			return h.fail(c, remote, req.Cmd, err), false
		}
		resp := okResponse(req.Cmd, countMessages(len(msgs)))
		resp.Messages = msgs
		h.event("historial", c.username, remote, "ok", "con="+req.With)
		return resp, false
	default:
		return h.fail(c, remote, req.Cmd, fmt.Errorf("%w: %q", models.ErrUnknownCommand, req.Cmd)), false
	}
}

func (h *Handler) login(c *connClient, remote string, req protocol.Request) protocol.Response {
	if c.username != "" {
		return errorResponse(req.Cmd, fmt.Errorf("%w: cierra la sesión actual antes de entrar con otro usuario", models.ErrAlreadyLoggedIn))
	}
	session, err := h.svc.Login(req.User, "", remote)
	if err != nil {
		if isUserError(err) {
			// Se deja constancia de qué usuario se intentó usar (entre comillas, para que un nombre
			// con caracteres raros no pueda falsear el registro).
			h.event(req.Cmd, "-", remote, "rechazado", fmt.Sprintf("usuario=%q: %v", req.User, err))
			return errorResponse(req.Cmd, err)
		}
		return h.fail(c, remote, req.Cmd, err)
	}
	c.username = session.Username()
	h.hub.register(c.username, c)
	h.event("login", c.username, remote, "ok", "")
	return okResponse(req.Cmd, "sesión iniciada como "+c.username)
}

func (h *Handler) logout(c *connClient, remote, cmd string) protocol.Response {
	if c.username == "" {
		return errorResponse(cmd, models.ErrNotLoggedIn)
	}
	name := c.username
	if err := h.svc.Logout(name); err != nil {
		return h.fail(c, remote, cmd, err)
	}
	h.hub.unregister(name, c)
	c.username = ""
	h.event("logout", name, remote, "ok", "")
	return okResponse(cmd, "sesión cerrada")
}

func (h *Handler) say(c *connClient, remote string, req protocol.Request) protocol.Response {
	msg, members, err := h.svc.SendToRoom(c.username, req.Room, req.Text)
	if err != nil {
		return h.fail(c, remote, req.Cmd, err)
	}
	delivered := h.hub.deliver(members, messageEvent(msg))
	h.event("mensaje", c.username, remote, "ok", fmt.Sprintf("sala=%s id=%d entregado_a=%d", msg.Room, msg.ID, delivered))
	return okResponse(req.Cmd, fmt.Sprintf("mensaje %d enviado a %s", msg.ID, msg.Room))
}

func (h *Handler) direct(c *connClient, remote string, req protocol.Request) protocol.Response {
	msg, err := h.svc.SendDirect(c.username, req.To, req.Text)
	if err != nil {
		return h.fail(c, remote, req.Cmd, err)
	}
	delivered := h.hub.deliver([]string{msg.From, msg.To}, messageEvent(msg))
	h.event("mensaje", c.username, remote, "ok", fmt.Sprintf("directo id=%d entregado_a=%d", msg.ID, delivered))
	return okResponse(req.Cmd, fmt.Sprintf("mensaje %d enviado a %s", msg.ID, msg.To))
}

// fail registra el error y construye la respuesta. Los errores de dominio se muestran tal cual;
// los internos (por ejemplo, un fallo de disco) se registran con detalle pero al cliente solo le
// llega un mensaje genérico, para no exponer rutas ni datos del servidor.
func (h *Handler) fail(c *connClient, remote, cmd string, err error) protocol.Response {
	if isUserError(err) {
		h.event(cmd, c.username, remote, "rechazado", err.Error())
		return errorResponse(cmd, err)
	}
	h.event(cmd, c.username, remote, "error_interno", err.Error())
	return protocol.Response{Type: protocol.TypeResponse, Cmd: cmd, Error: "no se pudo procesar la solicitud (error interno del servidor)"}
}

// event escribe una línea de registro. Nunca incluye el texto de los mensajes ni credenciales.
func (h *Handler) event(name, user, remote, result, detail string) {
	if user == "" {
		user = "-"
	}
	if detail != "" {
		h.logger.Printf("evento=%s usuario=%s origen=%s resultado=%s detalle=%q", name, user, remote, result, detail)
		return
	}
	h.logger.Printf("evento=%s usuario=%s origen=%s resultado=%s", name, user, remote, result)
}

func countMessages(n int) string {
	if n == 1 {
		return "1 mensaje"
	}
	return fmt.Sprintf("%d mensajes", n)
}

func isUserError(err error) bool {
	for _, target := range models.UserErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func okResponse(cmd, info string) protocol.Response {
	return protocol.Response{Type: protocol.TypeResponse, OK: true, Cmd: cmd, Info: info}
}

func errorResponse(cmd string, err error) protocol.Response {
	return protocol.Response{Type: protocol.TypeResponse, Cmd: cmd, Error: err.Error()}
}

func messageEvent(m models.Message) protocol.Response {
	return protocol.Response{Type: protocol.TypeMessage, OK: true, Message: &m}
}
