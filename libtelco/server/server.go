// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	"github.com/masyagin1998/SchoolServer/libtelco/push"
	api "github.com/masyagin1998/SchoolServer/libtelco/rest-api"

	"net/http"
	"runtime"
	// ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
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
		push:   push.NewPush(rest, logger),
		serv:   &http.Server{Addr: config.ServerAddr, Handler: rest.BindHandlers()},
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() error {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)

	/*
		// ТЕСТЫ.
		kek := ss.NewSession(&serv.config.Schools[0])
		if err := kek.Login(); err != nil {
			fmt.Println(err)
		}

		data, err := kek.GetParentInfoLetterData("11198")
		fmt.Println(data)
		if err != nil {
			fmt.Println(err)
		}

		if err := kek.Logout(); err != nil {
			fmt.Println(err)
		}
	*/

	defer func() {
		_ = serv.api.Db.Close()
		_ = serv.api.Redis.Close()
		_ = serv.api.Store.Close()
	}()

	// Подключить рассылку пушей
	// go serv.push.Run()

	return serv.serv.ListenAndServe()
}
