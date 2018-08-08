// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"fmt"
	"os"

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
		os.Exit(1)
	}
	data, _ := kek.GetWeekSchoolMarks("30.04.2018", "11198")
	fmt.Println(data)
	fmt.Println()
	data1, _ := kek.GetLessonDescription(241817, 13074, 3, "11198")
	fmt.Println(data1)
	fmt.Println()
	if err = kek.Logout(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	serv.api.BindHandlers()
	defer func() {
		_ = serv.api.Db.Close()
		_ = serv.api.Redis.Close()
		_ = serv.api.Store.Close()
	}()
	return http.ListenAndServe(serv.config.ServerAddr, context.ClearHandler(http.DefaultServeMux))
}
