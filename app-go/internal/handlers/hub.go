// Package handlers conecta la red (conexiones TCP) con los servicios de negocio.
package handlers

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

const writeTimeout = 5 * time.Second

// connClient es una conexión de un cliente. Varias goroutines pueden escribirle (la suya y las
// de otros clientes que le envían mensajes), por eso la escritura va protegida con un candado.
type connClient struct {
	conn     net.Conn
	mu       sync.Mutex
	enc      *json.Encoder
	username string // solo lo toca la goroutine propietaria de la conexión
}

func newConnClient(conn net.Conn) *connClient {
	return &connClient{conn: conn, enc: json.NewEncoder(conn)}
}

// Send escribe una respuesta como una línea JSON.
func (c *connClient) Send(resp protocol.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.enc.Encode(resp)
}

// Hub lleva el registro de qué usuario está conectado por qué conexión, para poder entregarle
// mensajes en tiempo real.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*connClient
}

// NewHub crea un Hub vacío.
func NewHub() *Hub {
	return &Hub{clients: make(map[string]*connClient)}
}

func (h *Hub) register(username string, c *connClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[username] = c
}

// unregister solo borra la entrada si todavía pertenece a esa conexión.
func (h *Hub) unregister(username string, c *connClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[username] == c {
		delete(h.clients, username)
	}
}

// deliver envía la respuesta a los usuarios indicados que estén conectados y devuelve cuántos la recibieron.
func (h *Hub) deliver(usernames []string, resp protocol.Response) int {
	h.mu.RLock()
	targets := make([]*connClient, 0, len(usernames))
	for _, name := range usernames {
		if c, ok := h.clients[name]; ok {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	delivered := 0
	for _, c := range targets {
		if err := c.Send(resp); err == nil {
			delivered++
		}
	}
	return delivered
}
