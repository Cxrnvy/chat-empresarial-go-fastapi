// Comando client: cliente de consola del chat empresarial.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/client"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:8080", "dirección IP y puerto del servidor")
	user := flag.String("user", "", "usuario con el que iniciar sesión automáticamente (opcional)")
	flag.Parse()

	c, err := client.Dial(*addr, 5*time.Second)
	if err != nil {
		return err
	}
	fmt.Printf("Conectado a %s. Escribe /help para ver los comandos.\n", *addr)
	return client.Run(c, os.Stdin, os.Stdout, *user)
}
