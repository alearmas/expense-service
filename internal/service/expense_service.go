package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

// ExpenseRepository define el contrato de persistencia que el servicio requiere.
// Cualquier tipo que implemente Save y FindAll puede usarse (real DynamoDB o mock de tests).
type ExpenseRepository interface {
	Save(ctx context.Context, expense *model.Expense) error
	FindAll(ctx context.Context) ([]*model.Expense, error)
}

// ExpenseService contiene la lógica de negocio para gastos
type ExpenseService struct {
	repo ExpenseRepository
}

// NewExpenseService crea una nueva instancia del servicio
func NewExpenseService(repo ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

// Register valida y persiste un nuevo gasto
func (s *ExpenseService) Register(ctx context.Context, req *model.CreateExpenseRequest) (*model.Expense, error) {
	if err := s.validate(req); err != nil {
		return nil, err
	}

	expense := &model.Expense{
		ExpenseID:     uuid.New().String(),
		Total:         req.Total,
		Category:      req.Category,
		Description:   req.Description,
		PaymentMethod: req.PaymentMethod,
		ExpenseDate:   req.ExpenseDate,
		Recipient:     req.Recipient,
		IsRecurring:   req.IsRecurring,
		Recurrence:    req.Recurrence,
	}

	slog.Info("📥 Registrando gasto", "expenseID", expense.ExpenseID, "category", expense.Category)

	if err := s.repo.Save(ctx, expense); err != nil {
		return nil, fmt.Errorf("error al guardar el gasto: %w", err)
	}

	slog.Info("✅ Gasto registrado exitosamente", "expenseID", expense.ExpenseID)
	return expense, nil
}

// GetAll retorna todos los gastos registrados
func (s *ExpenseService) GetAll(ctx context.Context) ([]*model.Expense, error) {
	slog.Info("🔎 Obteniendo todos los gastos")

	expenses, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener gastos: %w", err)
	}

	return expenses, nil
}

// validate aplica las reglas de negocio sobre el request
func (s *ExpenseService) validate(req *model.CreateExpenseRequest) error {
	if req.Total <= 0 {
		return &apperrors.ErrInvalidInput{Field: "total", Message: "debe ser mayor a 0"}
	}
	if req.Category == "" {
		return &apperrors.ErrInvalidInput{Field: "category", Message: "es obligatoria"}
	}
	if req.PaymentMethod == "" {
		return &apperrors.ErrInvalidInput{Field: "paymentMethod", Message: "es obligatorio"}
	}
	if req.ExpenseDate == "" {
		return &apperrors.ErrInvalidInput{Field: "expenseDate", Message: "es obligatoria"}
	}
	if req.Description == "" {
		return &apperrors.ErrInvalidInput{Field: "description", Message: "es obligatoria"}
	}
	return nil
}
