// Package services contiene la lógica de negocio del chat.
package services

import "github.com/Cxrnvy/chat-empresarial-go-fastapi/app-go/internal/models"

// Authenticator es el punto por donde, más adelante, se conectará el servicio de autenticación
// FastAPI (Fundamentos de Seguridad). El servidor de chat solo depende de esta interfaz, así que
// cambiar de implementación no obliga a tocar el resto del código.
type Authenticator interface {
	Authenticate(username, credential string) (*models.User, error)
}

// DirectoryAuthenticator es la implementación provisional del prototipo: solo identifica al usuario
// contra el directorio cargado desde users.json. No verifica contraseñas ni emite tokens, porque esa
// responsabilidad es del servicio FastAPI y no debe duplicarse dentro de Go.
type DirectoryAuthenticator struct {
	users map[string]*models.User
}

// NewDirectoryAuthenticator construye el directorio a partir de la lista de usuarios.
func NewDirectoryAuthenticator(users []*models.User) *DirectoryAuthenticator {
	d := &DirectoryAuthenticator{users: make(map[string]*models.User, len(users))}
	for _, u := range users {
		d.users[u.Username()] = u
	}
	return d
}

// Authenticate ignora la credencial a propósito (ver el comentario del tipo).
func (d *DirectoryAuthenticator) Authenticate(username, _ string) (*models.User, error) {
	if !models.ValidUsername(username) {
		return nil, models.ErrInvalidUsername
	}
	u, ok := d.users[username]
	if !ok {
		return nil, models.ErrUnknownUser
	}
	if !u.Active() {
		return nil, models.ErrInactiveUser
	}
	return u, nil
}
