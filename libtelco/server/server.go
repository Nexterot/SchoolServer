// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	api "SchoolServer/libtelco/rest-api"
	ss "SchoolServer/libtelco/sessions"
	"fmt"
	"net/http"
	"runtime"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config   *cp.Config
	sessions map[string]*ss.Session
	api      *api.RestAPI
	logger   *log.Logger
}

// NewServer создает новый сервер.
func NewServer(config *cp.Config, logger *log.Logger) *Server {
	sessions := make(map[string]*ss.Session)
	for _, schoolServer := range config.SchoolServers {
		sessions[schoolServer.Link] = ss.NewSession(&schoolServer)
	}
	serv := &Server{
		config:   config,
		sessions: sessions,
		api:      api.NewRestAPI(logger),
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() error {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)

	for _, session := range serv.sessions {
		if err := session.Login(); err != nil {
			serv.logger.Error("Error occured, while connecting to school server",
				"link", session.Serv.Link,
				"error", err)
		} else {
			fmt.Println("Session created")
		}
	}

	// Тест создания расписания.
	timeTable, err := serv.sessions[serv.config.SchoolServers[0].Link].GetDayTimeTable("17.04.2018")
	fmt.Println(timeTable)
	// Тест создания отчета.
	_, err = serv.sessions[serv.config.SchoolServers[0].Link].GetTotalMarkReport()
	if err != nil {
		fmt.Println(err)
	}

	// Подключаем handler'ы из RestAPI.
	serv.api.BindHandlers()
	return http.ListenAndServe(":8000", nil)
}
