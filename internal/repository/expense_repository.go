package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/alearmas/expense-service/internal/model"
)

// ExpenseRepository maneja las operaciones de DynamoDB para gastos
type ExpenseRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewExpenseRepository crea una nueva instancia del repositorio
func NewExpenseRepository(client *dynamodb.Client, tableName string) *ExpenseRepository {
	return &ExpenseRepository{client: client, tableName: tableName}
}

// Save persiste un gasto en DynamoDB
func (r *ExpenseRepository) Save(ctx context.Context, expense *model.Expense) error {
	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		return fmt.Errorf("error marshalling expense: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("error saving expense to DynamoDB: %w", err)
	}

	slog.Info("💾 Gasto guardado en DynamoDB", "expenseID", expense.ExpenseID)
	return nil
}

// FindAll retorna todos los gastos almacenados
func (r *ExpenseRepository) FindAll(ctx context.Context) ([]*model.Expense, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &r.tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning expenses from DynamoDB: %w", err)
	}

	expenses := make([]*model.Expense, 0, len(result.Items))
	for _, item := range result.Items {
		var expense model.Expense
		if err := attributevalue.UnmarshalMap(item, &expense); err != nil {
			return nil, fmt.Errorf("error unmarshalling expense: %w", err)
		}
		expenses = append(expenses, &expense)
	}

	slog.Info("📋 Gastos obtenidos de DynamoDB", "count", len(expenses))
	return expenses, nil
}
