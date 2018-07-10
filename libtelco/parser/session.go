// Copyright (C) 2018 Mikhail Masyagin

/*
Package parser - данный файл содержит в себе сессию парсера на нужном школьном сайте.
*/
package parser

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"fmt"
	"sync"
	"time"

	gr "github.com/levigross/grequests"
)

// session struct содержит в себе описание сессии к одному из школьных серверов.
type session struct {
	// Общая структура.
	sess     *gr.Session
	serv     *cp.SchoolServer
	ready    bool
	reload   chan struct{}
	ignTimer chan struct{}
	logger   *log.Logger
	mu       sync.Mutex
	// Для серверов первого типа.
	at  string
	ver string
}

// newSession создает новую сессию на базе информации о школьном сервере,
// к которому предстоит подключиться.
func newSession(server *cp.SchoolServer, logger *log.Logger) *session {
	return &session{
		sess:     nil,
		serv:     server,
		ready:    false,
		reload:   make(chan struct{}),
		ignTimer: make(chan struct{}),
		logger:   logger,
		mu:       sync.Mutex{},
	}
}

// timer раз в s.serv.Time секунд отправляет запрос на перезагрузку go-рутины.
func (s *session) timer() {
	time.Sleep(time.Duration(s.serv.Time) * time.Second)
	s.reload <- struct{}{}
	select {
	case <-s.ignTimer:
		return
	default:
		s.reload <- struct{}{}
	}
}

// startSession подключается к серверу и держит с ним соединение всё отведенное время.
// Как только время заканчивается (например, на 62.117.74.43 стоит убогое ограничение в 45 минут,
// мы заново коннектимся).
func (s *session) startSession(ch chan<- struct{}) {
	flag := false
	for {
		// Подключаемся к серверу.
		s.mu.Lock()
		s.ready = false
		if err := s.login(); err != nil {
			s.logger.Error("Error occured, while connecting to server",
				"Type", s.serv.Type,
				"Login", s.serv.Login,
				"Password", s.serv.Password,
				"error", err)
		} else {
			s.ready = true
			if !flag {
				s.logger.Info("New session was successfully created")
				flag = true
				ch <- struct{}{}
			} else {
				s.logger.Info("Session was successfully reloaded")
			}
			go s.timer()
		}
		s.mu.Unlock()
		// Ожидание требования перезагрузки.
		<-s.reload
	}
}

// TimeTable struct - расписание на N дней (N = 1, 2, ..., 7).
type TimeTable struct {
	Days []DayTimeTable `json:"days"`
}

// DayTimeTable struct - расписание на день.
type DayTimeTable struct {
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lesson"`
}

// Lesson struct - один урок.
type Lesson struct {
	Begin     string `json:"begin"`
	End       string `json:"end"`
	Name      string `json:"name"`
	ClassRoom string `json:"classroom"`
}

func (s *session) getTimeTable(date string, n int) (*TimeTable, error) {
	var err error
	var timeTable *TimeTable
	if (n < 1) || (n > 7) {
		err = fmt.Errorf("Invalid days number")
		return timeTable, err
	}
	timeTable = &TimeTable{
		Days: make([]DayTimeTable, 0, n),
	}
	for i := 0; i < n; i++ {
		day, err := s.getDayTimeTable(date)
		if err != nil {
			return timeTable, err
		}
		timeTable.Days = append(timeTable.Days, *day)
		date, err = incDate(date)
		if err != nil {
			return timeTable, err
		}
	}
	return timeTable, err
}

func incDate(date string) (string, error) {
	// Пока здесь пусто.
	return "", nil
}

func (s *session) getDayTimeTable(date string) (*DayTimeTable, error) {
	var err error
	var dayTimeTable *DayTimeTable
	switch s.serv.Type {
	case cp.FirstType:
		dayTimeTable, err = s.getDayTimeTableFirst(date)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.serv.Type)
	}
	return dayTimeTable, err
}

//func (s *session) getSchoolMarks(date string) *School
