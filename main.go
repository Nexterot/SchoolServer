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
		"Redis info", config.Redis,
		"Max allowed threads", config.MaxProcs,
		"Pool size", config.PoolSize,
		"Update Interval", config.UpdateInterval,
		"LogFile", config.LogFile,
	)
	// Вся информация о списке серверов.
	logger.Info("List of servers")
	for _, schoolServer := range config.SchoolServers {
		logger.Info("Server",
			"Type", schoolServer.Type,
			"Link", schoolServer.Link,
			"Login", schoolServer.Login,
			"Password", schoolServer.Password,
		)
	}
	// Запуск сервера.
	server := server.NewServer(config, logger)
	logger.Error("Fatal error occured, while running server", "error", server.Run())
}
