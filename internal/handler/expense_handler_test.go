package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/alearmas/expense-service/internal/handler"
	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ─────────────────────────────────────────────
// mockService implementa handler.ExpenseService
// ─────────────────────────────────────────────
type mockService struct {
	expense     *model.Expense
	expenses    []*model.Expense
	registerErr error
	getAllErr    error
}

func (m *mockService) Register(_ context.Context, _ *model.CreateExpenseRequest) (*model.Expense, error) {
	return m.expense, m.registerErr
}

func (m *mockService) GetAll(_ context.Context) ([]*model.Expense, error) {
	return m.expenses, m.getAllErr
}

// ─────────────────────────────────────
// helper: armar router de test
// ─────────────────────────────────────
func newRouter(svc handler.ExpenseService) *gin.Engine {
	r := gin.New()
	h := handler.NewExpenseHandler(svc)
	r.POST("/expenses", h.Register)
	r.GET("/expenses", h.GetAll)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("error al serializar body: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ─────────────────────────────────────
// POST /expenses — Register
// ─────────────────────────────────────

func TestRegisterHandler_ValidRequest_Returns201WithExpense(t *testing.T) {
	returnedExpense := &model.Expense{
		ExpenseID:     "test-uuid",
		Total:         500,
		Category:      model.CategoryStock,
		Description:   "Compra mercadería",
		PaymentMethod: model.PaymentTransfer,
		ExpenseDate:   "2026-03-04",
	}
	r := newRouter(&mockService{expense: returnedExpense})

	body := jsonBody(t, map[string]any{
		"total":         500,
		"category":      "STOCK",
		"description":   "Compra mercadería",
		"paymentMethod": "TRANSFER",
		"expenseDate":   "2026-03-04",
	})

	req := httptest.NewRequest(http.MethodPost, "/expenses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var got model.Expense
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("error al parsear response: %v", err)
	}
	if got.ExpenseID != "test-uuid" {
		t.Errorf("expenseID esperado 'test-uuid', obtuvo '%s'", got.ExpenseID)
	}
	if got.Total != 500 {
		t.Errorf("total esperado 500, obtuvo %v", got.Total)
	}
}

func TestRegisterHandler_InvalidJSON_Returns400(t *testing.T) {
	r := newRouter(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString("{not-json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuvo %d", w.Code)
	}
}

func TestRegisterHandler_ServiceValidationError_Returns400(t *testing.T) {
	svc := &mockService{
		registerErr: &apperrors.ErrInvalidInput{Field: "total", Message: "debe ser mayor a 0"},
	}
	r := newRouter(svc)

	body := jsonBody(t, map[string]any{
		"total": -1, "category": "STOCK",
		"description": "x", "paymentMethod": "CASH", "expenseDate": "2026-03-04",
	})
	req := httptest.NewRequest(http.MethodPost, "/expenses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400, obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == "" {
		t.Error("response debería contener campo 'error'")
	}
}

func TestRegisterHandler_ServiceNotFoundError_Returns404(t *testing.T) {
	svc := &mockService{
		registerErr: &apperrors.ErrNotFound{Resource: "Producto", ID: "xyz"},
	}
	r := newRouter(svc)

	body := jsonBody(t, map[string]any{
		"total": 100, "category": "STOCK",
		"description": "x", "paymentMethod": "CASH", "expenseDate": "2026-03-04",
	})
	req := httptest.NewRequest(http.MethodPost, "/expenses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperaba 404, obtuvo %d", w.Code)
	}
}

func TestRegisterHandler_ServiceInternalError_Returns500(t *testing.T) {
	svc := &mockService{registerErr: errors.New("error de DynamoDB")}
	r := newRouter(svc)

	body := jsonBody(t, map[string]any{
		"total": 100, "category": "STOCK",
		"description": "x", "paymentMethod": "CASH", "expenseDate": "2026-03-04",
	})
	req := httptest.NewRequest(http.MethodPost, "/expenses", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperaba 500, obtuvo %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "error interno del servidor" {
		t.Errorf("mensaje de error inesperado: %s", resp["error"])
	}
}

// ─────────────────────────────────────
// GET /expenses — GetAll
// ─────────────────────────────────────

func TestGetAllHandler_ReturnsExpenses_200(t *testing.T) {
	expenses := []*model.Expense{
		{ExpenseID: "a", Total: 100, Category: model.CategoryStock},
		{ExpenseID: "b", Total: 200, Category: model.CategoryRent},
	}
	r := newRouter(&mockService{expenses: expenses})

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuvo %d. Body: %s", w.Code, w.Body.String())
	}

	var got []*model.Expense
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("error al parsear response: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("esperaba 2 gastos, obtuvo %d", len(got))
	}
}

func TestGetAllHandler_EmptyList_Returns200WithEmptyArray(t *testing.T) {
	r := newRouter(&mockService{expenses: []*model.Expense{}})

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperaba 200, obtuvo %d", w.Code)
	}

	var got []*model.Expense
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("error al parsear response: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("esperaba array vacío, obtuvo %d elementos", len(got))
	}
}

func TestGetAllHandler_ServiceError_Returns500(t *testing.T) {
	r := newRouter(&mockService{getAllErr: errors.New("error de DynamoDB")})

	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperaba 500, obtuvo %d", w.Code)
	}
}
