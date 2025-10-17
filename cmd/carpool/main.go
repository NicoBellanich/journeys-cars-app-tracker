package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/controllers"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/docs"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/logger"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/services"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/storage"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/utils"
)

func main() {

	appLogger := logger.New("car-pooling-app")
	appLogger.Info("Starting car pooling service", map[string]interface{}{
		"version": "1.0.0",
	})

	storageType := utils.GetEnv("STORAGE_TYPE", "memory")
	appLogger.Info("Initializing storage", map[string]interface{}{
		"storage_type": storageType,
	})

	transactionFactory := storage.NewTransactionFactory(storageType)

	carPoolService := services.NewCarPool(transactionFactory)

	engine := gin.New()
	engine.Use(logger.GinMiddleware(appLogger))
	engine.Use(gin.Recovery())

	carPoolController := controllers.NewCarPool(carPoolService)

	wire(engine, carPoolController)

	// Serve OpenAPI and docs
	engine.GET("/openapi.yaml", func(ctx *gin.Context) {
		ctx.File("internal/docs/openapi.yaml")
	})
	engine.GET("/docs", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusOK, docs.SwaggerHTML)
	})

	ginMode := utils.GetEnv("GIN_MODE", gin.ReleaseMode)
	gin.SetMode(ginMode)

	host := utils.GetEnv("HOST", "0.0.0.0")
	port := utils.GetEnv("PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	server := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "graceful shutdown failed: %v\n", err)
		_ = server.Close()
	}
}

func wire(e *gin.Engine, c *controllers.CarPool) {
	e.GET("/status", c.GetStatus)
	e.Any("/cars", c.PutCars)
	e.Any("/journey", c.PostJourney)
	e.Any("/dropoff", c.PostDropoff)
	e.Any("/locate", c.PostLocate)
}
