// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	//"fmt"
	//"os"

	//ss "SchoolServer/libtelco/sessions"

	api "SchoolServer/libtelco/rest-api"

	"net/http"
	"runtime"

	"github.com/gorilla/context"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config *cp.Config
	api    *api.RestAPI
	logger *log.Logger
}

// NewServer создает новый сервер.
func NewServer(config *cp.Config, logger *log.Logger) *Server {
	serv := &Server{
		config: config,
		api:    api.NewRestAPI(logger, config),
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() error {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)

	serv.api.BindHandlers()
	defer func() {
		_ = serv.api.Db.Close()
		_ = serv.api.Redis.Close()
		_ = serv.api.Store.Close()
	}()
	return http.ListenAndServe(serv.config.ServerAddr, context.ClearHandler(http.DefaultServeMux))
}
