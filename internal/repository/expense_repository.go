package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/alearmas/expense-service/internal/model"
	"github.com/alearmas/expense-service/pkg/apperrors"
)

type ExpenseRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewExpenseRepository(client *dynamodb.Client, tableName string) *ExpenseRepository {
	return &ExpenseRepository{client: client, tableName: tableName}
}

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

func (r *ExpenseRepository) FindAll(ctx context.Context) ([]*model.Expense, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{TableName: &r.tableName})
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

func (r *ExpenseRepository) FindByID(ctx context.Context, expenseID string) (*model.Expense, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"expenseID": expenseID})
	if err != nil {
		return nil, fmt.Errorf("error marshalling key: %w", err)
	}
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting expense from DynamoDB: %w", err)
	}
	if result.Item == nil {
		return nil, &apperrors.ErrNotFound{Resource: "expense", ID: expenseID}
	}
	var expense model.Expense
	if err := attributevalue.UnmarshalMap(result.Item, &expense); err != nil {
		return nil, fmt.Errorf("error unmarshalling expense: %w", err)
	}
	return &expense, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		return fmt.Errorf("error marshalling expense: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(expenseID)"),
	})
	if err != nil {
		return fmt.Errorf("error updating expense in DynamoDB: %w", err)
	}
	slog.Info("✏️ Gasto actualizado en DynamoDB", "expenseID", expense.ExpenseID)
	return nil
}

func (r *ExpenseRepository) Delete(ctx context.Context, expenseID string) error {
	key, err := attributevalue.MarshalMap(map[string]string{"expenseID": expenseID})
	if err != nil {
		return fmt.Errorf("error marshalling key: %w", err)
	}
	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           &r.tableName,
		Key:                 key,
		ConditionExpression: aws.String("attribute_exists(expenseID)"),
	})
	if err != nil {
		return fmt.Errorf("error deleting expense from DynamoDB: %w", err)
	}
	slog.Info("🗑️ Gasto eliminado de DynamoDB", "expenseID", expenseID)
	return nil
}
