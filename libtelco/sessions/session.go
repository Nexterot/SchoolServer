// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"sync"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	gr "github.com/levigross/grequests"
)

// Session struct содержит в себе описание сессии к одному из школьных серверов.
type Session struct {
	dt.Session
	mu sync.Mutex
}

// NewSession создает новую сессию на базе информации о школьном сервере,
// к которому предстоит подключиться.
func NewSession(server *cp.School) *Session {
	s := &Session{}
	s.Sess = gr.NewSession(nil)
	s.Serv = server
	return s
}
