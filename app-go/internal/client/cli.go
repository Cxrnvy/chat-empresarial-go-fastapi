package client

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

// printer escribe líneas en la salida desde varias goroutines sin mezclarlas.
type printer struct {
	mu  sync.Mutex
	out io.Writer
}

func (p *printer) println(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintln(p.out, s)
}

// Run ejecuta la interfaz de consola: lee comandos de in, los envía al servidor y muestra en out
// tanto las respuestas como los mensajes que llegan en tiempo real. Si user no está vacío, inicia
// sesión automáticamente. Funciona igual con teclado o con entrada redirigida (para demostraciones).
func Run(c *Client, in io.Reader, out io.Writer, user string) error {
	p := &printer{out: out}
	done := make(chan error, 1)
	go func() {
		done <- c.Listen(func(r protocol.Response) { p.println(FormatResponse(r)) })
	}()

	if user != "" {
		if err := c.Send(protocol.Request{Cmd: protocol.CmdLogin, User: user}); err != nil {
			return err
		}
	}

	lines := make(chan string)
	go func() {
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	room := ""
	for {
		select {
		case err := <-done:
			if err != nil {
				p.println("ERROR: " + err.Error())
			}
			return err
		case line, ok := <-lines:
			if !ok {
				return finish(c, done)
			}
			cmd, err := ParseCommand(line, room)
			if err != nil {
				p.println("ERROR: " + err.Error())
				continue
			}
			switch cmd.Action {
			case ActionHelp:
				p.println(HelpText)
			case ActionSetRoom:
				room = cmd.Room
				p.println("Sala actual: " + room)
			case ActionQuit:
				return finish(c, done)
			case ActionSend:
				if err := c.Send(cmd.Request); err != nil {
					p.println("ERROR: " + err.Error())
					return err
				}
			}
		}
	}
}

// finish espera las respuestas pendientes, se despide del servidor y cierra la conexión.
func finish(c *Client, done <-chan error) error {
	deadline := time.Now().Add(2 * time.Second)
	for c.Pending() > 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	_ = c.Send(protocol.Request{Cmd: protocol.CmdQuit})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	return c.Close()
}
