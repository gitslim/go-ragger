package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func SetupLogger() {
	// Настраиваем логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func SetupFileLogger(fileName string) {
	// Открываем файл для логирования
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Errorf("failed to open log file: %w", err))
	}

	// Настраиваем логгер
	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

}
