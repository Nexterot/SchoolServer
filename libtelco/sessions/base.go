// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе сессии на серверах школ.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions/session"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/type-01"

	gr "github.com/levigross/grequests"
	"github.com/pkg/errors"
)

// Session struct содержит в себе описание сессии к одному из школьных серверов.
type Session struct {
	Base *ss.Session
}

// NewSession создает новую сессию на базе информации о школьном сервере,
// к которому предстоит подключиться.
func NewSession(server *cp.School) *Session {
	return &Session{
		Base: &ss.Session{
			Sess: gr.NewSession(nil),
			Serv: server,
		},
	}
}

/*
Вход в систему.
*/

// Login логинится к серверу и создает очередную сессию.
func (s *Session) Login() error {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	switch s.Base.Serv.Type {
	case cp.FirstType:
		err = t01.Login(s.Base)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return errors.Wrap(err, "from Login")
}

/*
Выход из системы.
*/

// Logout выходит с сервера.
func (s *Session) Logout() error {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	switch s.Base.Serv.Type {
	case cp.FirstType:
		err = t01.Logout(s.Base)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return errors.Wrap(err, "from Logout")
}

/*
Получение списка детей.
*/

// GetChildrenMap получает мапу детей в их ID.
func (s *Session) GetChildrenMap() error {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	switch s.Base.Serv.Type {
	case cp.FirstType:
		err = t01.GetChildrenMap(s.Base)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return errors.Wrap(err, "from GetChildrenMap")
}

/*
Получение списка предметов.
*/

// GetLessonsMap возвращает список пар мапу предметов в их ID.
func (s *Session) GetLessonsMap(studentID string) (*dt.LessonsMap, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var lessonsMap *dt.LessonsMap
	switch s.Base.Serv.Type {
	case cp.FirstType:
		lessonsMap, err = t01.GetLessonsMap(s.Base, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return lessonsMap, errors.Wrap(err, "from GetLessonsMap")
}
