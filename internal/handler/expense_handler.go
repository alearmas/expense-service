package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

// ExpenseService define el contrato de negocio que el handler requiere.
type ExpenseService interface {
	Register(ctx context.Context, req *model.CreateExpenseRequest) (*model.Expense, error)
	GetAll(ctx context.Context) ([]*model.Expense, error)
	Update(ctx context.Context, expenseID string, req *model.UpdateExpenseRequest) (*model.Expense, error)
	Delete(ctx context.Context, expenseID string) error
}

// ExpenseHandler maneja las peticiones HTTP para gastos
type ExpenseHandler struct {
	service ExpenseService
}

// NewExpenseHandler crea una nueva instancia del handler
func NewExpenseHandler(s ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: s}
}

// Register godoc
// POST /expenses
func (h *ExpenseHandler) Register(c *gin.Context) {
	var req model.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	slog.Info("📩 Recibida solicitud de nuevo gasto")

	expense, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, expense)
}

// GetAll godoc
// GET /expenses
func (h *ExpenseHandler) GetAll(c *gin.Context) {
	slog.Info("📩 Recibida solicitud para obtener todos los gastos")

	expenses, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, expenses)
}

// Update godoc
// PUT /expenses/:id
func (h *ExpenseHandler) Update(c *gin.Context) {
	expenseID := c.Param("id")
	if expenseID == "" {
		c.JSON(http.StatusBadRequest, errorResponse("expenseID requerido"))
		return
	}

	var req model.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	slog.Info("📩 Recibida solicitud de actualización de gasto", "expenseID", expenseID)

	expense, err := h.service.Update(c.Request.Context(), expenseID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, expense)
}

// Delete godoc
// DELETE /expenses/:id
func (h *ExpenseHandler) Delete(c *gin.Context) {
	expenseID := c.Param("id")
	if expenseID == "" {
		c.JSON(http.StatusBadRequest, errorResponse("expenseID requerido"))
		return
	}

	slog.Info("📩 Recibida solicitud de eliminación de gasto", "expenseID", expenseID)

	if err := h.service.Delete(c.Request.Context(), expenseID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleServiceError mapea errores de negocio a respuestas HTTP
func handleServiceError(c *gin.Context, err error) {
	var invalidInput *apperrors.ErrInvalidInput
	var notFound *apperrors.ErrNotFound

	switch {
	case errors.As(err, &invalidInput):
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	case errors.As(err, &notFound):
		c.JSON(http.StatusNotFound, errorResponse(err.Error()))
	default:
		slog.Error("❌ Error interno", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse("error interno del servidor"))
	}
}

func errorResponse(message string) gin.H {
	return gin.H{"error": message}
}
