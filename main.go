// Copyright (C) 2018 Mikhail Masyagin

package main

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"SchoolServer/libtelco/parser"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
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
// - инициализацию SQL базы данных;
// - инициализацию InMemory базы данных;
func init() {
	if config, err = cp.ReadConfig(); err != nil {
		os.Exit(1)
	}
	if logger, err = log.NewLogger(config.LogFile); err != nil {
		os.Exit(1)
	}
	/*
		if sqlDB, err = gorm.Open("postgres", config.Postgres); err != nil {
			os.Exit(1)
		}
	*/
}

func main() {
	logger.Info("SchoolServer V1.0 is running",
		"Server address", config.ServerAddr,
		"Postgres info", config.Postgres,
		"Redis info", config.Redis,
		"Max allowed threads", config.MaxProcs,
		"Pool size", config.PoolSize,
		"Update Interval", config.UpdateInterval,
		"LogFile", config.LogFile,
	)
	logger.Info("List of servers")
	for _, schoolServer := range config.SchoolServers {
		logger.Info("Server",
			"Type", schoolServer.Type,
			"Link", schoolServer.Link,
			"Login", schoolServer.Login,
			"Password", schoolServer.Password,
		)
	}

	timeTable := parser.NewTimeTable()
	if err = timeTable.ParseSchoolServer(&config.SchoolServers[0]); err != nil {
		logger.Error("Error occured, while running server", "error", err)
	}
}
