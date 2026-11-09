// Package storage se encarga de leer y guardar los datos del prototipo.
package storage

import "github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"

// Store separa a los servicios de la forma concreta de guardar los datos.
// Hoy se implementa con archivos JSON (FileStore); más adelante podría ser una base de datos
// sin tocar la lógica de negocio.
type Store interface {
	LoadUsers() ([]*models.User, error)
	LoadRooms() ([]*models.Room, error)
	NextMessageID() int
	AppendMessage(msg models.Message) error
	Messages() ([]models.Message, error)
}
