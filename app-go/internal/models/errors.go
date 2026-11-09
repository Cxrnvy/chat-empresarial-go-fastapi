// Package models define las entidades principales del chat y sus reglas de validación.
package models

import "errors"

// Errores de dominio. Se comparan con errors.Is y sus mensajes se muestran tal cual al usuario,
// por eso están escritos para que se entiendan sin conocer el código.
var (
	ErrEmptyMessage    = errors.New("el mensaje está vacío")
	ErrMessageTooLong  = errors.New("el mensaje supera los 1000 caracteres")
	ErrInvalidUsername = errors.New("nombre de usuario inválido: use de 3 a 20 letras minúsculas, números o guion bajo")
	ErrInvalidRoom     = errors.New("nombre de sala inválido: use de 3 a 30 letras minúsculas, números, guion o guion bajo")
	ErrInvalidRole     = errors.New("rol inválido: use user o admin")
	ErrUnknownUser     = errors.New("el usuario no existe")
	ErrInactiveUser    = errors.New("la cuenta está desactivada")
	ErrUnknownRoom     = errors.New("la sala no existe")
	ErrNotMember       = errors.New("no eres miembro de esa sala")
	ErrNotLoggedIn     = errors.New("primero inicia sesión con /login <usuario>")
	ErrAlreadyLoggedIn = errors.New("ya hay una sesión activa para ese usuario")
	ErrSelfMessage     = errors.New("no puedes enviarte un mensaje directo a ti mismo")
	ErrInvalidRequest  = errors.New("solicitud inválida")
	ErrUnknownCommand  = errors.New("comando desconocido")
)

// UserErrors agrupa los errores que el servidor puede mostrar al cliente sin riesgo.
// Cualquier otro error se considera interno y se reporta con un mensaje genérico.
var UserErrors = []error{
	ErrEmptyMessage, ErrMessageTooLong, ErrInvalidUsername, ErrInvalidRoom, ErrInvalidRole,
	ErrUnknownUser, ErrInactiveUser, ErrUnknownRoom, ErrNotMember, ErrNotLoggedIn,
	ErrAlreadyLoggedIn, ErrSelfMessage, ErrInvalidRequest, ErrUnknownCommand,
}
