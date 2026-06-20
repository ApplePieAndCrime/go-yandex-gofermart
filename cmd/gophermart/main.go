package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/accrual"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/auth"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/database"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/handler"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/service"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/worker"
	"go.uber.org/zap"
)

func main() {
	var runAddr, databaseURI, accrualAddr string
	flag.StringVar(&runAddr, "a", "", "адрес и порт запуска сервиса (например, localhost:8080)")
	flag.StringVar(&databaseURI, "d", "", "строка подключения к PostgreSQL")
	flag.StringVar(&accrualAddr, "r", "", "адрес системы расчёта начислений")
	flag.Parse()

	if runAddr == "" {
		runAddr = os.Getenv("RUN_ADDRESS")
	}
	if databaseURI == "" {
		databaseURI = os.Getenv("DATABASE_URI")
	}
	if accrualAddr == "" {
		accrualAddr = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}

	if runAddr == "" {
		runAddr = "localhost:8080"
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	if err := auth.InitJWT(); err != nil {
		sugar.Fatalf("JWT init failed: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, databaseURI)
	if err != nil {
		sugar.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := repository.NewRepository(pool)
	services := service.NewService(repo, accrualAddr, sugar)
	h := handler.NewHandler(services, sugar)

	router := h.InitRoutes()

	srv := &http.Server{
		Addr:         runAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Фоновый воркер (обновление статусов заказов)
	accrualClient := accrual.NewClient(accrualAddr)

	// Запускаем воркер
	worker := worker.NewWorker(repo, accrualClient, sugar, 5*time.Second)
	go worker.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sugar.Infof("Starting server on %s", runAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("server failed: %v", err)
		}
	}()

	<-stop
	sugar.Info("Shutting down gracefully...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		sugar.Errorf("server shutdown error: %v", err)
	}

	worker.Stop()

	sugar.Info("Server stopped")
}
