package apperrors

import "fmt"

// ErrNotFound se devuelve cuando un recurso no existe
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s con ID '%s' no encontrado", e.Resource, e.ID)
}

// ErrInvalidInput se devuelve cuando el input de negocio es inválido
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("campo '%s': %s", e.Field, e.Message)
}
