// Copyright (C) 2018 Mikhail Masyagin

package main

import (
	"fmt"
	"os"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
)

var (
	// Конфиг сервера.
	config *cp.Config
	// Логгер.
	logger *log.Logger
	// Стандартная ошибка.
	err error
)

// init производит:
// - чтение конфигурационных файлов;
// - создание логгера;
func init() {
	if config, err = cp.ReadConfig(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if logger, err = log.NewLogger(config.LogFile); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

func main() {
	// Вся информация о конфиге.
	logger.Info("SchoolServer V0.1 is running",
		"Server address", config.ServerAddr,
		"Postgres info", config.Postgres,
		"Max allowed threads", config.MaxProcs,
		"LogFile", config.LogFile,
	)
	// Вся информация о списке серверов.
	logger.Info("List of schools")
	for _, school := range config.Schools {
		logger.Info("School",
			"Name", school.Name,
			"Type", school.Type,
			"Link", school.Link,
			"Permission", school.Permission,
		)
	}

	// ТЕСТЫ.
	kek := ss.NewSession(&config.Schools[0])
	if err := kek.Login(); err != nil {
		fmt.Println(err)
	}

	data, err := kek.GetAnnouncements("1", "127.0.0.1:1488")
	if err != nil {
		fmt.Println(err)
	}
	for i, post := range data.Posts {
		fmt.Printf("%v)\n", i)
		fmt.Println(post.Title)
		fmt.Println(post.FileLink)
		fmt.Println(post.FileID)
		fmt.Println(post.FileName)
	}

	if err := kek.Logout(); err != nil {
		fmt.Println(err)
	}
	os.Exit(1)

	// Запуск сервера.
	/*
		server := server.NewServer(config, logger)
		server.Run()
	*/
}
