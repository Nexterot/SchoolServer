// Copyright (C) 2018 Mikhail Masyagin

/*
Package parser - данный файл содержит пул "воркеров" парсера.
*/
package parser

import (
	"sync"

	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
)

// Task interface - описание интрефейса задачи.
type Task interface {
	Execute()
}

// Pool struct содержит в себе "воркеры" парсеров.
// - кол-во "воркеров", доступных пользователю (всего Size + N пулов,
// где N - количество серверов).
type Pool struct {
	size     int
	tasks    chan Task
	logger   *log.Logger
	wg       sync.WaitGroup
	sessions map[string]*session
}

// NewPool создает новый пул парсеров.
func NewPool(size int, servers []cp.SchoolServer, logger *log.Logger) *Pool {
	// Запуск сессий всех парсеров.
	ch := make(chan struct{})
	sessionMap := make(map[string]*session)
	for _, server := range servers {
		sessionMap[server.Link] = newSession(&server, logger)
		go sessionMap[server.Link].startSession(ch)
		<-ch
	}
	// Создание пула парсеров.
	pool := &Pool{
		size:     size,
		tasks:    make(chan Task, 128),
		logger:   logger,
		sessions: sessionMap,
	}
	for i := 0; i < size+len(servers); i++ {
		pool.wg.Add(1)
		go pool.newWorker(ch, i)
		<-ch
	}
	logger.Info("New Parsers Pool was successfully created")
	pool.sessions["62.117.74.43"].getDayTimeTable("06.03.2018")
	return pool
}

// newWorker создает новый "воркер".
func (pool *Pool) newWorker(ch chan<- struct{}, i int) {
	pool.logger.Info("New Worker was succesfully created",
		"ID", i)
	ch <- struct{}{}
	defer func() {
		pool.wg.Done()
		pool.logger.Info("Worker was suddenly closed",
			"ID", i)
	}()
	for {
		select {
		// Если есть задача, то ее нужно обработать
		case task, ok := <-pool.tasks:
			if !ok {
				return
			}
			task.Execute()
		}
	}
}

// Close закрывает канал задач пула "воркеров".
func (pool *Pool) Close() {
	close(pool.tasks)
}

// Wait ожидает завершения работы всех "воркеров".
func (pool *Pool) Wait() {
	pool.wg.Wait()
}

// Execute запускает очередную задачу на выполнение.
func (pool *Pool) Execute(task Task) {
	pool.tasks <- task
}
