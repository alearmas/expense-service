package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/internal/service"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

// ─────────────────────────────────────────────
// mockRepo implementa service.ExpenseRepository
// ─────────────────────────────────────────────
type mockRepo struct {
	saved    *model.Expense
	expenses []*model.Expense
	saveErr  error
	findErr  error
}

func (m *mockRepo) Save(_ context.Context, e *model.Expense) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = e
	return nil
}

func (m *mockRepo) FindAll(_ context.Context) ([]*model.Expense, error) {
	return m.expenses, m.findErr
}

// ─────────────────────────────────────
// helpers
// ─────────────────────────────────────
func validRequest() *model.CreateExpenseRequest {
	return &model.CreateExpenseRequest{
		Total:         500.0,
		Category:      model.CategoryStock,
		Description:   "Compra de mercadería",
		PaymentMethod: model.PaymentTransfer,
		ExpenseDate:   "2026-03-04",
	}
}

// ─────────────────────────────────────
// Register — happy path
// ─────────────────────────────────────

func TestRegister_ValidRequest_ReturnsExpense(t *testing.T) {
	mock := &mockRepo{}
	svc := service.NewExpenseService(mock)

	expense, err := svc.Register(context.Background(), validRequest())

	if err != nil {
		t.Fatalf("esperaba nil error, obtuvo: %v", err)
	}
	if expense == nil {
		t.Fatal("esperaba un expense, obtuvo nil")
	}
	if expense.ExpenseID == "" {
		t.Error("ExpenseID no debería estar vacío")
	}
	if expense.Total != 500.0 {
		t.Errorf("Total esperado 500.0, obtuvo %.2f", expense.Total)
	}
	if mock.saved == nil {
		t.Error("el repositorio debería haber guardado el expense")
	}
}

func TestRegister_ValidRequest_SetsAllFields(t *testing.T) {
	req := &model.CreateExpenseRequest{
		Total:         300.0,
		Category:      model.CategoryRent,
		Description:   "Alquiler marzo",
		PaymentMethod: model.PaymentTransfer,
		ExpenseDate:   "2026-03-01",
		Recipient:     "Inmobiliaria XYZ",
		IsRecurring:   true,
		Recurrence: &model.RecurrenceDetails{
			Type:      model.RecurrenceMensual,
			StartDate: "2026-01-01",
		},
	}

	mock := &mockRepo{}
	svc := service.NewExpenseService(mock)

	expense, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("esperaba nil error, obtuvo: %v", err)
	}

	if expense.Category != model.CategoryRent {
		t.Errorf("Category esperado %s, obtuvo %s", model.CategoryRent, expense.Category)
	}
	if expense.Recipient != "Inmobiliaria XYZ" {
		t.Errorf("Recipient esperado 'Inmobiliaria XYZ', obtuvo '%s'", expense.Recipient)
	}
	if !expense.IsRecurring {
		t.Error("IsRecurring debería ser true")
	}
	if expense.Recurrence == nil || expense.Recurrence.Type != model.RecurrenceMensual {
		t.Error("Recurrence no se mapeó correctamente")
	}
}

// ─────────────────────────────────────
// Register — validaciones
// ─────────────────────────────────────

func TestRegister_NegativeTotal_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.Total = -100

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "total")
}

func TestRegister_ZeroTotal_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.Total = 0

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "total")
}

func TestRegister_EmptyCategory_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.Category = ""

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "category")
}

func TestRegister_EmptyPaymentMethod_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.PaymentMethod = ""

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "paymentMethod")
}

func TestRegister_EmptyExpenseDate_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.ExpenseDate = ""

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "expenseDate")
}

func TestRegister_EmptyDescription_ReturnsInvalidInputError(t *testing.T) {
	req := validRequest()
	req.Description = ""

	svc := service.NewExpenseService(&mockRepo{})
	_, err := svc.Register(context.Background(), req)

	assertInvalidInput(t, err, "description")
}

// ─────────────────────────────────────
// Register — errores de repo
// ─────────────────────────────────────

func TestRegister_RepoError_PropagatesError(t *testing.T) {
	repoErr := errors.New("error de DynamoDB")
	mock := &mockRepo{saveErr: repoErr}
	svc := service.NewExpenseService(mock)

	_, err := svc.Register(context.Background(), validRequest())

	if err == nil {
		t.Fatal("esperaba error, obtuvo nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("esperaba que el error wrapeara repoErr, obtuvo: %v", err)
	}
}

// ─────────────────────────────────────
// GetAll
// ─────────────────────────────────────

func TestGetAll_ReturnsAllExpenses(t *testing.T) {
	expected := []*model.Expense{
		{ExpenseID: "abc", Total: 100, Category: model.CategoryStock},
		{ExpenseID: "def", Total: 200, Category: model.CategoryRent},
	}
	mock := &mockRepo{expenses: expected}
	svc := service.NewExpenseService(mock)

	result, err := svc.GetAll(context.Background())

	if err != nil {
		t.Fatalf("esperaba nil error, obtuvo: %v", err)
	}
	if len(result) != len(expected) {
		t.Errorf("esperaba %d gastos, obtuvo %d", len(expected), len(result))
	}
	if result[0].ExpenseID != "abc" {
		t.Errorf("primer expense esperado 'abc', obtuvo '%s'", result[0].ExpenseID)
	}
}

func TestGetAll_EmptyList_ReturnsEmptySlice(t *testing.T) {
	mock := &mockRepo{expenses: []*model.Expense{}}
	svc := service.NewExpenseService(mock)

	result, err := svc.GetAll(context.Background())

	if err != nil {
		t.Fatalf("esperaba nil error, obtuvo: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("esperaba slice vacío, obtuvo %d elementos", len(result))
	}
}

func TestGetAll_RepoError_PropagatesError(t *testing.T) {
	repoErr := errors.New("error de DynamoDB")
	mock := &mockRepo{findErr: repoErr}
	svc := service.NewExpenseService(mock)

	_, err := svc.GetAll(context.Background())

	if err == nil {
		t.Fatal("esperaba error, obtuvo nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("esperaba que el error wrapeara repoErr, obtuvo: %v", err)
	}
}

// ─────────────────────────────────────
// helper de assertions
// ─────────────────────────────────────

func assertInvalidInput(t *testing.T, err error, expectedField string) {
	t.Helper()
	var invalidErr *apperrors.ErrInvalidInput
	if !errors.As(err, &invalidErr) {
		t.Fatalf("esperaba *apperrors.ErrInvalidInput, obtuvo: %T: %v", err, err)
	}
	if invalidErr.Field != expectedField {
		t.Errorf("campo esperado '%s', obtuvo '%s'", expectedField, invalidErr.Field)
	}
}
