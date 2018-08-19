// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе функции для отправки и чтения сообщений на форуме.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/forum"
	"github.com/pkg/errors"
)

// GetForumThemesList возвращает список тем форума.
func (s *Session) GetForumThemesList(page string) (*dt.ForumThemesList, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var forumThemesList *dt.ForumThemesList
	switch s.Serv.Type {
	case cp.FirstType:
		forumThemesList, err = t01.GetForumThemesList(&s.Session, page)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return forumThemesList, errors.Wrap(err, "from GetForumThemesList")
}

// GetForumThemeMessages возвращает список всех сообщений одной темы форума.
func (s *Session) GetForumThemeMessages(TID, page, pageSize string) (*dt.ForumThemeMessages, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var forumThemeMessages *dt.ForumThemeMessages
	switch s.Serv.Type {
	case cp.FirstType:
		forumThemeMessages, err = t01.GetForumThemeMessages(&s.Session, TID, page, pageSize)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return forumThemeMessages, errors.Wrap(err, "from GetForumTheme")
}

// CreateForumTheme создаёт новую тему на форуме.
func (s *Session) CreateForumTheme(page, name, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.CreateForumTheme(&s.Session, page, name, message)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from GetForumTheme")
}

// CreateForumThemeMessage создаёт новое сообщение в теме на форуме.
func (s *Session) CreateForumThemeMessage(page, message, TID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.CreateForumThemeMessage(&s.Session, page, message, TID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from GetForumTheme")
}
