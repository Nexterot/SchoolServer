// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"fmt"
	"time"

	ss "SchoolServer/libtelco/sessions"

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
	// Тесты.
	kek := ss.NewSession(&serv.config.Schools[0])

	err := kek.Login()
	if err != nil {
		fmt.Println(err)
	}

	err = kek.GetChildrenMap()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(kek.Base.Children)
	data, err := kek.GetWeekSchoolMarks("11.10.2017", "11198")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	time.Sleep(time.Minute * time.Duration(25))
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	data, err = kek.GetWeekSchoolMarks("11.10.2017", "11198")
	if err != nil {
		fmt.Println(err)
	}

	err = kek.Logout()
	if err != nil {
		fmt.Println(err)
	}
	serv.api.BindHandlers()
	return http.ListenAndServe(serv.config.ServerAddr, context.ClearHandler(http.DefaultServeMux))
}
