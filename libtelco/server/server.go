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
	/*
		_, err = kek.GetAverageMarkReport("19.03.2018", "25.05.2018", "T")
		if err != nil {
			fmt.Println(err)
		}

		_, err = kek.GetAverageMarkReportDyn("04.09.2017", "29.06.2018", "T")
		if err != nil {
			fmt.Println(err)
		}

		_, err = kek.GetStudentGradeReport("01.09.2017", "31.08.2018", "13076")
		if err != nil {
			fmt.Println(err)
		}
	*/

	_, err = kek.GetStudentTotalReport("01.09.2017", "31.08.2018")
	if err != nil {
		fmt.Println(err)
	}

	// Подключаем handler'ы из RestAPI.
	serv.api.BindHandlers()
	return http.ListenAndServe(serv.config.ServerAddr, nil)
}
