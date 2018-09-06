// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	"github.com/masyagin1998/SchoolServer/libtelco/push"
	api "github.com/masyagin1998/SchoolServer/libtelco/rest-api"

	"net/http"
	"runtime"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config *cp.Config
	api    *api.RestAPI
	logger *log.Logger
	push   *push.Push
	serv   *http.Server
}

// NewServer создает новый сервер.
func NewServer(config *cp.Config, logger *log.Logger) *Server {
	rest := api.NewRestAPI(logger, config)
	serv := &Server{
		config: config,
		api:    rest,
		logger: logger,
		push:   push.NewPush(rest, logger),
		serv:   &http.Server{Addr: config.ServerAddr, Handler: rest.BindHandlers()},
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)

	// Подключить рассылку пушей
	// go serv.push.Run()

	defer func() {
		_ = serv.api.Db.Close()
		_ = serv.api.Redis.Close()
		_ = serv.api.Store.Close()
	}()

	go func() {
		if err := serv.serv.ListenAndServe(); err != http.ErrServerClosed {
			serv.logger.Error("Fatal error occured, while running server", "error", err)
		}
	}()

	serv.gracefulShutdown()
}

// gracefulShutdown безопасно выключает сервер.
func (serv *Server) gracefulShutdown() {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	serv.logger.Info("Got signal", "signal", "SIGTERM")
	serv.logger.Info("Gracefull shutdown was successfully started")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()

	if err := serv.serv.Shutdown(ctx); err != nil {
		serv.logger.Error("Error occured, while closing server", "error", err)
	} else {
		serv.logger.Info("Server was successfully shutdowned")
	}
}
