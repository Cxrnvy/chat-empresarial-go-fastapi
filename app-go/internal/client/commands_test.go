package client

import (
	"testing"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/protocol"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		room    string
		action  Action
		cmd     string
		wantErr bool
	}{
		{"línea vacía", "   ", "", ActionNone, "", false},
		{"login", "/login ariel", "", ActionSend, protocol.CmdLogin, false},
		{"login sin usuario", "/login", "", ActionSend, "", true},
		{"login con espacios", "/login ariel jorge", "", ActionSend, "", true},
		{"texto sin sala", "hola", "", ActionSend, "", true},
		{"texto con sala", "hola", "general", ActionSend, protocol.CmdSay, false},
		{"say", "/say general hola a todos", "", ActionSend, protocol.CmdSay, false},
		{"say sin texto lo decide el servidor", "/say general", "", ActionSend, protocol.CmdSay, false},
		{"dm", "/dm jorge hola", "", ActionSend, protocol.CmdDM, false},
		{"dm sin destinatario", "/dm", "", ActionSend, "", true},
		{"history con sala y límite", "/history general 5", "", ActionSend, protocol.CmdHistory, false},
		{"history usa la sala actual", "/history 5", "general", ActionSend, protocol.CmdHistory, false},
		{"history sin sala", "/history", "", ActionSend, "", true},
		{"dmhistory", "/dmhistory jorge 3", "", ActionSend, protocol.CmdDMHistory, false},
		{"dmhistory con límite inválido", "/dmhistory jorge tres", "", ActionSend, "", true},
		{"room es local", "/room general", "", ActionSetRoom, "", false},
		{"ayuda", "/help", "", ActionHelp, "", false},
		{"salir", "/quit", "", ActionQuit, "", false},
		{"comando desconocido", "/volar", "", ActionSend, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCommand(tc.line, tc.room)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, se esperaba error: %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got.Action != tc.action || got.Request.Cmd != tc.cmd {
				t.Fatalf("resultado = %+v, se esperaba acción %v y comando %q", got, tc.action, tc.cmd)
			}
		})
	}
}

func TestSayKeepsTextForServerValidation(t *testing.T) {
	got, err := ParseCommand("/say general", "")
	if err != nil || got.Request.Text != "" || got.Request.Room != "general" {
		t.Fatalf("el cliente debe enviar el mensaje vacío al servidor para que lo valide: %+v, %v", got, err)
	}
}
