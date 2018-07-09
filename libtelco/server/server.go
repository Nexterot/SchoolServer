// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"SchoolServer/libtelco/parser"
	"runtime"
	"time"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config *cp.Config
	parser *parser.Pool
	logger *log.Logger
}

// NewServer создает новый сервер.
func NewServer(config *cp.Config, logger *log.Logger) *Server {
	serv := &Server{
		config: config,
		logger: logger,
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() error {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)
	// Запускаем пул.
	serv.parser = parser.NewPool(serv.config.PoolSize,
		serv.config.SchoolServers,
		serv.logger)
	time.Sleep(time.Duration(100) * time.Second)
	return nil
}
