package model

// ExpenseCategory representa la categoría del gasto
type ExpenseCategory string

const (
	CategoryStock     ExpenseCategory = "STOCK"
	CategoryRent      ExpenseCategory = "RENT"
	CategoryMarketing ExpenseCategory = "MARKETING"
	CategorySalaries  ExpenseCategory = "SALARIES"
	CategoryServices  ExpenseCategory = "SERVICES"
	CategoryTaxes     ExpenseCategory = "TAXES"
	CategoryOthers    ExpenseCategory = "OTHERS"
)

// PaymentMethod representa el método de pago
type PaymentMethod string

const (
	PaymentCash        PaymentMethod = "CASH"
	PaymentTransfer    PaymentMethod = "TRANSFER"
	PaymentCreditCard  PaymentMethod = "CREDIT_CARD"
	PaymentDebitCard   PaymentMethod = "DEBIT_CARD"
	PaymentMercadoPago PaymentMethod = "MERCADO_PAGO"
	PaymentOther       PaymentMethod = "OTHER"
)

// RecurrenceType representa la frecuencia de un gasto recurrente
type RecurrenceType string

const (
	RecurrenceMensual    RecurrenceType = "MENSUAL"
	RecurrenceBimestral  RecurrenceType = "BIMESTRAL"
	RecurrenceTrimestral RecurrenceType = "TRIMESTRAL"
	RecurrenceAnual      RecurrenceType = "ANUAL"
)

// RecurrenceDetails contiene la información de recurrencia de un gasto
type RecurrenceDetails struct {
	Type      RecurrenceType `json:"type"      dynamodbav:"type"`
	StartDate string         `json:"startDate" dynamodbav:"startDate"` // YYYY-MM-DD
	EndDate   string         `json:"endDate"   dynamodbav:"endDate"`   // YYYY-MM-DD
}

// Expense representa un gasto registrado
type Expense struct {
	ExpenseID     string             `json:"expenseID"               dynamodbav:"expenseID"`
	Total         float64            `json:"total"                   dynamodbav:"total"`
	Category      ExpenseCategory    `json:"category"                dynamodbav:"category"`
	Description   string             `json:"description"             dynamodbav:"description"`
	PaymentMethod PaymentMethod      `json:"paymentMethod"           dynamodbav:"paymentMethod"`
	ExpenseDate   string             `json:"expenseDate"             dynamodbav:"expenseDate"` // YYYY-MM-DD
	Recipient     string             `json:"recipient,omitempty"     dynamodbav:"recipient,omitempty"`
	IsRecurring   bool               `json:"isRecurring"             dynamodbav:"isRecurring"`
	Recurrence    *RecurrenceDetails `json:"recurrence,omitempty"    dynamodbav:"recurrence,omitempty"`
}

// CreateExpenseRequest es el body esperado para registrar un gasto
type CreateExpenseRequest struct {
	Total         float64            `json:"total"          binding:"required,gt=0"`
	Category      ExpenseCategory    `json:"category"       binding:"required"`
	Description   string             `json:"description"    binding:"required"`
	PaymentMethod PaymentMethod      `json:"paymentMethod"  binding:"required"`
	ExpenseDate   string             `json:"expenseDate"    binding:"required"`
	Recipient     string             `json:"recipient,omitempty"`
	IsRecurring   bool               `json:"isRecurring"`
	Recurrence    *RecurrenceDetails `json:"recurrence,omitempty"`
}
