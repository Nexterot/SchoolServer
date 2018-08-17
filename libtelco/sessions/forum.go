// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе функции для отправки и чтения сообщений на форуме.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/type-01"
	"github.com/pkg/errors"
)

// GetForumThemesList возвращает список тем форума.
func (s *Session) GetForumThemesList(page string) (*dt.ForumThemesList, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	var forumThemesList *dt.ForumThemesList
	switch s.Base.Serv.Type {
	case cp.FirstType:
		forumThemesList, err = t01.GetForumThemesList(s.Base, page)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return forumThemesList, errors.Wrap(err, "from GetEmailsList")
}
