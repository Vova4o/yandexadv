package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/internal/server/handler"
	"github.com/vova4o/yandexadv/internal/server/middleware"
	"github.com/vova4o/yandexadv/internal/server/service"
	"github.com/vova4o/yandexadv/internal/server/storage"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	config := flags.NewConfig()

	logger, err := logger.NewLogger("info", config.ServerLogFile)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Build information",
		zap.String("version", buildVersion),
		zap.String("date", buildDate),
		zap.String("commit", buildCommit),
	)

	middle := middleware.New(logger, config.SecretKey)

	stor := storage.Init(config, logger)

	service := service.New(stor, logger)

	router := handler.New(service, middle, config.CryptoPath)
	router.RegisterRoutes()

	// Создание канала для получения сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		if err := router.StartServer(config.ServerAddress); err != nil {
			logger.Error("Failed to start server", zap.Error(err))
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	go func() {
		logger.Info("Starting ppof server on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			logger.Error("Failed to start pprof server", zap.Error(err))
			log.Fatalf("Failed to start pprof server: %v", err)
		}
	}()

	// Ожидание сигнала завершения работы
	<-stop

	// Создание контекста с тайм-аутом для завершения работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := stor.Stop(); err != nil {
		logger.Error("Failed to stop storage", zap.Error(err))
	}

	// Логирование завершения работы сервера
	logger.Info("Shutting down server...")

	// Завершение работы сервера
	if err := router.StopServer(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
