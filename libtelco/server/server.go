// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит "бизнес-логику" сервера,
является своеобразным ядром.
*/
package server

import (
	"log"

	cp "SchoolServer/libtelco/config-parser"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Server struct содержит основную информацию о сервере:
// - адрес сервера;
// - sql базы данных;
// - inMemory базу данных;
// - логгер;
// - размер пула;
// - время обновления (в секундах);
// - параметры школьных серверов;
type Server struct {
	serverAddr     string
	sqlDB          *gorm.DB
	inMemoryDB     *redis.Client
	logger         *log.Logger
	poolSize       int
	updateInterval int
	schoolServers  []cp.SchoolServer
}

// NewServer создакт новый сервер.
func NewServer(serverAddr string,
	sqlDB *gorm.DB, inMemoryDB *redis.Client,
	logger *log.Logger,
	poolSize int,
	updateInterval int,
	schoolServers []cp.SchoolServer) *Server {
	return &Server{serverAddr,
		sqlDB, inMemoryDB,
		logger,
		poolSize,
		updateInterval,
		schoolServers}
}

// Run стартует сервер.
func (server *Server) Run() {

}

// updateSQLDB создает, если надо, таблицы базы данных
// и запускает автоматическое обновление расписания.
func (server *Server) updateSQLDB() {

}
