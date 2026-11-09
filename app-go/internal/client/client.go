// Package client contiene la conexión del cliente con el servidor y la interfaz de consola.
package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

// ErrConnectionLost indica que el servidor cerró la conexión o dejó de responder.
var ErrConnectionLost = errors.New("se perdió la conexión con el servidor")

// Client es la conexión TCP de un cliente.
type Client struct {
	conn    net.Conn
	enc     *json.Encoder
	pending int64 // solicitudes enviadas que todavía no tienen respuesta
	closed  int32
}

// Dial abre la conexión con el servidor. Si falla, el error explica qué dirección se intentó.
func Dial(addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar con el servidor en %s: %w", addr, err)
	}
	return &Client{conn: conn, enc: json.NewEncoder(conn)}, nil
}

// Send envía una solicitud al servidor.
func (c *Client) Send(req protocol.Request) error {
	atomic.AddInt64(&c.pending, 1)
	if err := c.enc.Encode(req); err != nil {
		atomic.AddInt64(&c.pending, -1)
		return fmt.Errorf("%w: %v", ErrConnectionLost, err)
	}
	return nil
}

// Listen lee lo que envía el servidor y se lo pasa a handle, hasta que la conexión se cierre.
// Devuelve nil si fue el propio cliente quien cerró, y ErrConnectionLost en cualquier otro caso.
func (c *Client) Listen(handle func(protocol.Response)) error {
	scanner := bufio.NewScanner(c.conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var resp protocol.Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue // una línea ilegible del servidor no debe tumbar al cliente
		}
		if resp.Type == protocol.TypeResponse {
			atomic.AddInt64(&c.pending, -1)
		}
		handle(resp)
	}
	if atomic.LoadInt32(&c.closed) == 1 {
		return nil
	}
	return ErrConnectionLost
}

// Pending devuelve cuántas solicitudes esperan respuesta.
func (c *Client) Pending() int64 { return atomic.LoadInt64(&c.pending) }

// Close cierra la conexión.
func (c *Client) Close() error {
	atomic.StoreInt32(&c.closed, 1)
	return c.conn.Close()
}
