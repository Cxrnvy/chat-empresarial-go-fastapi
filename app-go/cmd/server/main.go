// Comando server: servidor del chat empresarial.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/handlers"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/services"
	"github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:8080", "dirección IP y puerto en los que escucha el servidor")
	dataDir := flag.String("data", "data", "directorio con users.json, rooms.json y messages.jsonl")
	flag.Parse()

	store, err := storage.NewFileStore(*dataDir)
	if err != nil {
		return err
	}
	svc, err := services.NewChatService(store, nil, nil)
	if err != nil {
		return err
	}
	handler := handlers.NewHandler(svc, log.New(os.Stdout, "", log.LstdFlags))

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("no se pudo escuchar en %s: %w", *addr, err)
	}
	log.Printf("servidor escuchando en %s (datos en %s)", ln.Addr(), *dataDir)

	// Al recibir Ctrl+C se cierra el listener y el bucle de aceptación termina limpiamente.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("apagando el servidor")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("error al aceptar una conexión: %v", err)
			continue
		}
		go handler.ServeConn(conn)
	}
}
