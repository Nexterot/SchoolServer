// Copyright (C) 2018 Mikhail Masyagin

package main

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"SchoolServer/libtelco/server"
	"os"
)

var (
	// Конфиг сервера.
	config *cp.Config
	// Логгер.
	logger *log.Logger
	/*
		// ORM-Объект PostgreSQL
		sqlDB *gorm.DB
		// Объект Redis.
		inMemoryDB *redis.Client
	*/
	// Стандартная ошибка.
	err error
)

// init производит:
// - чтение конфигурационных файлов;
// - создание логгера;
func init() {
	if config, err = cp.ReadConfig(); err != nil {
		os.Exit(1)
	}
	if logger, err = log.NewLogger(config.LogFile); err != nil {
		os.Exit(1)
	}

}

func main() {
	// Вся информация о конфиге.
	logger.Info("SchoolServer V0.1 is running",
		"Server address", config.ServerAddr,
		"Postgres info", config.Postgres,
		"Max allowed threads", config.MaxProcs,
		"Update Interval", config.UpdateInterval,
		"LogFile", config.LogFile,
	)
	// Вся информация о списке серверов.
	logger.Info("List of schools")
	for _, school := range config.Schools {
		logger.Info("School",
			"Name", school.Name,
			"Type", school.Type,
			"Link", school.Link,
			"Time", school.Time,
			"Permission", school.Permission,
		)
	}

	// Запуск сервера.
	server := server.NewServer(config, logger)
	logger.Error("Fatal error occured, while running server", "error", server.Run())
}
