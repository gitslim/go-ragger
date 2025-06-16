package main

import (
	"github.com/gitslim/go-ragger/internal/app"
	"github.com/gitslim/go-ragger/internal/logger"
)

func main() {
	// Настройка логгера
	logger.SetupLogger()

	// Запуск приложения
	app.RunServerApp()
}
