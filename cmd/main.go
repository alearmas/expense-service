package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	appconfig "github.com/alearmas/expense-service/internal/config"
	"github.com/alearmas/expense-service/internal/handler"
	"github.com/alearmas/expense-service/internal/repository"
	"github.com/alearmas/expense-service/internal/service"
)

var ginLambda *ginadapter.GinLambda

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	tableName := os.Getenv("EXPENSES_TABLE_NAME")
	if tableName == "" {
		tableName = "expenses_table"
	}

	dynamoClient, err := appconfig.NewDynamoDBClient(context.Background())
	if err != nil {
		slog.Error("❌ Error iniciando cliente DynamoDB", "error", err)
		os.Exit(1)
	}

	repo := repository.NewExpenseRepository(dynamoClient, tableName)
	svc := service.NewExpenseService(repo)
	h := handler.NewExpenseHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	expenses := r.Group("/expenses")
	{
		expenses.POST("", h.Register)
		expenses.GET("", h.GetAll)
	}

	ginLambda = ginadapter.New(r)
	slog.Info("🚀 expense-service inicializado correctamente")
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
