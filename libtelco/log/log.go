// Copyright (C) 2018 Mikhail Masyagin
/*
Package log содержит "ООП-обертку" над популярным логгером mgutz/logxi.
Использование обертки вызвано необходимостью иметь единый интерфейс для:
- записи логов в консоль;
- записи логов в файл;
- отсутсвия записи логов;
*/
package log

import (
	"io"
	"os"
	"time"

	"github.com/mgutz/logxi/v1"
)

// Logger struct содержит основную информацию о логгере:
// - должен ли он вообще писать логи;
// - пишет ли он логи в консоль или в файл;
// - файл назначения (если есть);
type Logger struct {
	useLog bool
	logger log.Logger
	file   *os.File
}

// NewLogger создает новый логгер.
func NewLogger(config string) (logger *Logger, err error) {
	logger = &Logger{}
	switch config {
	case "stdout":
		logger.useLog = true
		logger.logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "SchoolServer")
	case "stderr":
		logger.useLog = true
		logger.logger = log.NewLogger(log.NewConcurrentWriter(os.Stderr), "SchoolServer")
	case "":
	default:
		logger.useLog = true
		fileName := config + "/log_from_" + time.Now().Format("02_01_2006_15h_04m_05s") + ".txt"
		logFile, err := os.Create(fileName)
		if err != nil {
			return nil, err
		}
		logger.logger = log.NewLogger(log.NewConcurrentWriter(io.Writer(logFile)), "SchoolServer")
	}
	return
}

// CloseLogger закрывает файл записи логов, если тот существует.
func (logger *Logger) CloseLogger() {
	if logger.file != nil {
		if err := logger.file.Close(); err != nil {
			_ = logger.logger.Error("Error occured while closing log file", "error", err)
		}
	}
}

// Info логгирует важную информацию.
func (logger *Logger) Info(msg string, arg ...interface{}) {
	if logger.useLog {
		logger.logger.Info(msg, arg...)
	}
}

// Error логгирует ошибки.
func (logger *Logger) Error(msg string, arg ...interface{}) {
	if logger.useLog {
		_ = logger.logger.Error(msg, arg...)
	}
}

// Fatal логгирует ошибки и завершает выполнение программы.
func (logger *Logger) Fatal(msg string, arg ...interface{}) {
	if logger.useLog {
		logger.logger.Fatal(msg, arg...)
	}
}
