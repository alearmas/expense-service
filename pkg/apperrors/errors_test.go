package apperrors_test

import (
	"errors"
	"testing"

	"github.com/alearmas/expense-service/pkg/apperrors"
)

// ─────────────────────────────────────
// ErrInvalidInput
// ─────────────────────────────────────

func TestErrInvalidInput_ErrorMessage(t *testing.T) {
	err := &apperrors.ErrInvalidInput{Field: "total", Message: "debe ser mayor a 0"}
	expected := "campo 'total': debe ser mayor a 0"
	if err.Error() != expected {
		t.Errorf("esperaba %q, obtuvo %q", expected, err.Error())
	}
}

func TestErrInvalidInput_DifferentFields(t *testing.T) {
	cases := []struct {
		field   string
		message string
		want    string
	}{
		{"category", "es obligatoria", "campo 'category': es obligatoria"},
		{"paymentMethod", "es obligatorio", "campo 'paymentMethod': es obligatorio"},
		{"expenseDate", "es obligatoria", "campo 'expenseDate': es obligatoria"},
		{"description", "es obligatoria", "campo 'description': es obligatoria"},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			err := &apperrors.ErrInvalidInput{Field: tc.field, Message: tc.message}
			if err.Error() != tc.want {
				t.Errorf("esperaba %q, obtuvo %q", tc.want, err.Error())
			}
		})
	}
}

func TestErrInvalidInput_ImplementsErrorInterface(t *testing.T) {
	var err error = &apperrors.ErrInvalidInput{Field: "x", Message: "y"}
	if err == nil {
		t.Error("ErrInvalidInput debería implementar error")
	}
}

func TestErrInvalidInput_ErrorsAs(t *testing.T) {
	original := &apperrors.ErrInvalidInput{Field: "total", Message: "debe ser mayor a 0"}
	wrapped := errors.New("wrap: " + original.Error())

	// errors.As sobre el tipo directamente
	var target *apperrors.ErrInvalidInput
	if errors.As(original, &target) {
		if target.Field != "total" {
			t.Errorf("field esperado 'total', obtuvo '%s'", target.Field)
		}
	}

	// wrapped string no debería hacer As
	var target2 *apperrors.ErrInvalidInput
	if errors.As(wrapped, &target2) {
		t.Error("errors.As no debería matchear un error wrapeado como string")
	}
}

// ─────────────────────────────────────
// ErrNotFound
// ─────────────────────────────────────

func TestErrNotFound_ErrorMessage(t *testing.T) {
	err := &apperrors.ErrNotFound{Resource: "Gasto", ID: "abc-123"}
	expected := "Gasto con ID 'abc-123' no encontrado"
	if err.Error() != expected {
		t.Errorf("esperaba %q, obtuvo %q", expected, err.Error())
	}
}

func TestErrNotFound_DifferentResources(t *testing.T) {
	cases := []struct {
		resource string
		id       string
		want     string
	}{
		{"Gasto", "uuid-1", "Gasto con ID 'uuid-1' no encontrado"},
		{"Producto", "sku-99", "Producto con ID 'sku-99' no encontrado"},
	}

	for _, tc := range cases {
		t.Run(tc.resource, func(t *testing.T) {
			err := &apperrors.ErrNotFound{Resource: tc.resource, ID: tc.id}
			if err.Error() != tc.want {
				t.Errorf("esperaba %q, obtuvo %q", tc.want, err.Error())
			}
		})
	}
}

func TestErrNotFound_ImplementsErrorInterface(t *testing.T) {
	var err error = &apperrors.ErrNotFound{Resource: "X", ID: "1"}
	if err == nil {
		t.Error("ErrNotFound debería implementar error")
	}
}

func TestErrNotFound_ErrorsAs(t *testing.T) {
	original := &apperrors.ErrNotFound{Resource: "Gasto", ID: "123"}

	var target *apperrors.ErrNotFound
	if !errors.As(original, &target) {
		t.Fatal("errors.As debería matchear ErrNotFound")
	}
	if target.Resource != "Gasto" {
		t.Errorf("Resource esperado 'Gasto', obtuvo '%s'", target.Resource)
	}
	if target.ID != "123" {
		t.Errorf("ID esperado '123', obtuvo '%s'", target.ID)
	}
}
