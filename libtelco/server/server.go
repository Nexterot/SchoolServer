// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"SchoolServer/libtelco/parser"
	"SchoolServer/libtelco/rest-api"
	"net/http"
	"runtime"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config  *cp.Config
	parser  *parser.Pool
	restapi *restapi.RestAPI
	logger  *log.Logger
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

	// Подключаем handler'ы из RestAPI.
	serv.restapi = restapi.NewRestAPI(serv.logger)
	serv.restapi.BindHandlers()
	// Запускаем гуся, работяги.
	return http.ListenAndServe(":8000", nil)
}
