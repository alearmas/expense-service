package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

type ExpenseRepository interface {
	Save(ctx context.Context, expense *model.Expense) error
	FindAll(ctx context.Context) ([]*model.Expense, error)
	FindByID(ctx context.Context, expenseID string) (*model.Expense, error)
	Update(ctx context.Context, expense *model.Expense) error
	Delete(ctx context.Context, expenseID string) error
}

type ExpenseService struct {
	repo ExpenseRepository
}

func NewExpenseService(repo ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

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

func (s *ExpenseService) GetAll(ctx context.Context) ([]*model.Expense, error) {
	slog.Info("🔎 Obteniendo todos los gastos")
	expenses, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener gastos: %w", err)
	}
	return expenses, nil
}

func (s *ExpenseService) Update(ctx context.Context, expenseID string, req *model.UpdateExpenseRequest) (*model.Expense, error) {
	existing, err := s.repo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}
	existing.Total = req.Total
	existing.Category = req.Category
	existing.Description = req.Description
	existing.PaymentMethod = req.PaymentMethod
	existing.ExpenseDate = req.ExpenseDate
	existing.Recipient = req.Recipient
	existing.IsRecurring = req.IsRecurring
	existing.Recurrence = req.Recurrence

	slog.Info("✏️ Actualizando gasto", "expenseID", expenseID)
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("error al actualizar el gasto: %w", err)
	}
	slog.Info("✅ Gasto actualizado exitosamente", "expenseID", expenseID)
	return existing, nil
}

func (s *ExpenseService) Delete(ctx context.Context, expenseID string) error {
	if _, err := s.repo.FindByID(ctx, expenseID); err != nil {
		return err
	}
	slog.Info("🗑️ Eliminando gasto", "expenseID", expenseID)
	if err := s.repo.Delete(ctx, expenseID); err != nil {
		return fmt.Errorf("error al eliminar el gasto: %w", err)
	}
	slog.Info("✅ Gasto eliminado exitosamente", "expenseID", expenseID)
	return nil
}

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
