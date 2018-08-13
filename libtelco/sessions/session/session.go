// Copyright (C) 2018 Mikhail Masyagin

/*
Package session содержит в себе структуру сессии.
*/
package session

import (
	"sync"
	"time"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"

	gr "github.com/levigross/grequests"
)

// Type является enum'ом для типа сессии:
type Type int

const (
	// Undefined - значение по умолчанию.
	Undefined Type = iota
	// Parent - значение для родителя.
	Parent
	// Child - значение для ребенка.
	Child
)

// Session struct содержит в себе описание сессии к одному из школьных серверов.
type Session struct {
	// Общая структура.
	Sess        *gr.Session
	Serv        *cp.School
	MU          sync.Mutex
	LastRequest time.Time
	// Тип: родитель или ученик.
	Type Type
	// Только для родителей с 2-мя и более детьми.
	Children map[string]Student
	// Для учеников, а также для родителей с одним ребенком.
	Child Student
	// Для серверов первого типа.
	AT  string
	VER string
}

// Student struct содержит в себе информацию о ребенке.
type Student struct {
	SID  string
	CLID string
}
