package handlers

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/services"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/storage"
)

// failingStore simula un fallo de disco al guardar mensajes.
type failingStore struct{ storage.Store }

func (failingStore) AppendMessage(models.Message) error {
	return errors.New("disco lleno en /ruta/secreta/messages.jsonl")
}

func seededStore(t *testing.T) storage.Store {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"users.json": `[{"username":"cesar","role":"admin","active":true},
			{"username":"ariel","role":"user","active":true},
			{"username":"jorge","role":"user","active":true}]`,
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
	return store
}

// startServer levanta un servidor real en un puerto libre de la máquina local.
func startServer(t *testing.T, store storage.Store) (string, *services.ChatService) {
	t.Helper()
	svc, err := services.NewChatService(store, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(svc, log.New(io.Discard, "", 0))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go h.ServeConn(conn)
		}
	}()
	return ln.Addr().String(), svc
}

type testClient struct {
	t    *testing.T
	conn net.Conn
	sc   *bufio.Scanner
}

func connect(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, conn: conn, sc: bufio.NewScanner(conn)}
}

func (c *testClient) sendRaw(line string) {
	c.t.Helper()
	if _, err := c.conn.Write([]byte(line + "\n")); err != nil {
		c.t.Fatal(err)
	}
}

func (c *testClient) send(req protocol.Request) {
	c.t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		c.t.Fatal(err)
	}
	c.sendRaw(string(data))
}

// next lee una línea del servidor, con un plazo para que una prueba rota no se quede colgada.
func (c *testClient) next() protocol.Response {
	c.t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if !c.sc.Scan() {
		c.t.Fatalf("no llegó respuesta del servidor: %v", c.sc.Err())
	}
	var r protocol.Response
	if err := json.Unmarshal(c.sc.Bytes(), &r); err != nil {
		c.t.Fatalf("respuesta ilegible: %v", err)
	}
	return r
}

// expect lee hasta encontrar una respuesta del tipo pedido (los mensajes en tiempo real pueden llegar antes).
func (c *testClient) expect(typ string) protocol.Response {
	c.t.Helper()
	for i := 0; i < 5; i++ {
		if r := c.next(); r.Type == typ {
			return r
		}
	}
	c.t.Fatalf("no llegó una respuesta de tipo %q", typ)
	return protocol.Response{}
}

func (c *testClient) login(user string) {
	c.t.Helper()
	c.send(protocol.Request{Cmd: protocol.CmdLogin, User: user})
	if r := c.expect(protocol.TypeResponse); !r.OK {
		c.t.Fatalf("login de %s falló: %s", user, r.Error)
	}
}

func TestRealTimeDelivery(t *testing.T) {
	addr, _ := startServer(t, seededStore(t))
	ariel, jorge := connect(t, addr), connect(t, addr)
	ariel.login("ariel")
	jorge.login("jorge")

	ariel.send(protocol.Request{Cmd: protocol.CmdSay, Room: "general", Text: "hola equipo"})

	push := jorge.expect(protocol.TypeMessage)
	if push.Message == nil || push.Message.From != "ariel" || push.Message.Text != "hola equipo" {
		t.Fatalf("jorge no recibió el mensaje en tiempo real: %+v", push)
	}
	if r := ariel.expect(protocol.TypeResponse); !r.OK {
		t.Fatalf("ariel debía recibir confirmación: %+v", r)
	}
}

func TestDirectMessageOnlyReachesParticipants(t *testing.T) {
	addr, _ := startServer(t, seededStore(t))
	ariel, jorge, cesar := connect(t, addr), connect(t, addr), connect(t, addr)
	ariel.login("ariel")
	jorge.login("jorge")
	cesar.login("cesar")

	ariel.send(protocol.Request{Cmd: protocol.CmdDM, To: "jorge", Text: "reunión a las 3"})
	if m := jorge.expect(protocol.TypeMessage); m.Message.Text != "reunión a las 3" || !m.Message.IsDirect() {
		t.Fatalf("jorge no recibió el mensaje directo: %+v", m)
	}
	// cesar pide algo que sí tiene respuesta: no debe haber recibido el mensaje privado antes.
	cesar.send(protocol.Request{Cmd: protocol.CmdWho})
	if r := cesar.next(); r.Type != protocol.TypeResponse {
		t.Fatalf("cesar recibió un mensaje que no era para él: %+v", r)
	}
}

func TestValidationErrorsAreReported(t *testing.T) {
	addr, _ := startServer(t, seededStore(t))
	c := connect(t, addr)

	cases := []struct {
		name string
		line string
		want string
	}{
		{"JSON mal formado", `esto no es json`, "solicitud inválida"},
		{"comando desconocido", `{"cmd":"volar"}`, "comando desconocido"},
		{"enviar sin sesión", `{"cmd":"say","room":"general","text":"hola"}`, "inicia sesión"},
		{"usuario inexistente", `{"cmd":"login","user":"fantasma"}`, "no existe"},
		{"usuario con formato inválido", `{"cmd":"login","user":"Mal Nombre"}`, "inválido"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.sendRaw(tc.line)
			r := c.next()
			if r.OK || !strings.Contains(r.Error, tc.want) {
				t.Fatalf("respuesta = %+v, se esperaba un error que contenga %q", r, tc.want)
			}
		})
	}

	c.login("ariel")
	after := []struct {
		name string
		req  protocol.Request
		want string
	}{
		{"mensaje vacío", protocol.Request{Cmd: protocol.CmdSay, Room: "general", Text: "   "}, "vacío"},
		{"sala inexistente", protocol.Request{Cmd: protocol.CmdSay, Room: "nada", Text: "hola"}, "no existe"},
		{"destinatario inexistente", protocol.Request{Cmd: protocol.CmdDM, To: "fantasma", Text: "hola"}, "no existe"},
		{"segunda sesión en la misma conexión", protocol.Request{Cmd: protocol.CmdLogin, User: "jorge"}, "ya hay una sesión"},
	}
	for _, tc := range after {
		t.Run(tc.name, func(t *testing.T) {
			c.send(tc.req)
			r := c.next()
			if r.OK || !strings.Contains(r.Error, tc.want) {
				t.Fatalf("respuesta = %+v, se esperaba un error que contenga %q", r, tc.want)
			}
		})
	}
}

func TestHistoryOverNetwork(t *testing.T) {
	addr, _ := startServer(t, seededStore(t))
	ariel, jorge := connect(t, addr), connect(t, addr)
	ariel.login("ariel")
	jorge.login("jorge")
	for _, text := range []string{"uno", "dos"} {
		ariel.send(protocol.Request{Cmd: protocol.CmdSay, Room: "general", Text: text})
		ariel.expect(protocol.TypeResponse)
	}
	jorge.send(protocol.Request{Cmd: protocol.CmdHistory, Room: "general"})
	r := jorge.expect(protocol.TypeResponse)
	if !r.OK || len(r.Messages) != 2 || r.Messages[0].Text != "uno" {
		t.Fatalf("historial incorrecto: %+v", r)
	}
	jorge.send(protocol.Request{Cmd: protocol.CmdHistory, Room: "proyectos"})
	if r := jorge.expect(protocol.TypeResponse); r.OK || !strings.Contains(r.Error, "no eres miembro") {
		t.Fatalf("jorge no debía ver el historial de proyectos: %+v", r)
	}
}

func TestDisconnectReleasesSession(t *testing.T) {
	addr, svc := startServer(t, seededStore(t))
	first := connect(t, addr)
	first.login("ariel")
	first.conn.Close() // el cliente se cae sin hacer logout

	deadline := time.Now().Add(2 * time.Second)
	for len(svc.OnlineUsers()) > 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := svc.OnlineUsers(); len(got) != 0 {
		t.Fatalf("la sesión no se liberó al desconectarse: %v", got)
	}
	second := connect(t, addr)
	second.login("ariel") // debe poder volver a entrar
}

func TestInternalErrorsDoNotLeakDetails(t *testing.T) {
	addr, _ := startServer(t, failingStore{seededStore(t)})
	c := connect(t, addr)
	c.login("ariel")
	c.send(protocol.Request{Cmd: protocol.CmdSay, Room: "general", Text: "hola"})
	r := c.next()
	if r.OK || strings.Contains(r.Error, "/ruta/secreta") || !strings.Contains(r.Error, "error interno") {
		t.Fatalf("el error interno debía ser genérico: %+v", r)
	}
}
