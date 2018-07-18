// Copyright (C) 2018 Mikhail Masyagin

/*
Package session содержит в себе структуру сессии.
*/
package session

import (
	cp "SchoolServer/libtelco/config-parser"
	"sync"
	"time"

	gr "github.com/levigross/grequests"
)

// Type является enum'ом для типа сессии:
type Type int

const (
	// Undefined - значение по умолчанию.
	Undefined Type = iota
	// Parent - значение для родителя.
	Parent
	// Student - значение для ученика.
	Student
)

// Session struct содержит в себе описание сессии к одному из школьных серверов.
type Session struct {
	// Общая структура.
	Sess        *gr.Session
	Serv        *cp.School
	MU          sync.Mutex
	LastRequest time.Time
	Type        Type
	// Только для родителей.
	ChildrenIDS map[string]string
	// Для серверов первого типа.
	AT  string
	VER string
}
