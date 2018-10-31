// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/announcements"
	"github.com/pkg/errors"
)

func (s *Session) GetAnnouncements(schooldID, serverAddr string) (*dt.Posts, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var posts *dt.Posts
	switch s.Serv.Type {
	case cp.FirstType:
		posts, err = t01.GetAnnouncements(&s.Session, schooldID, serverAddr)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return posts, errors.Wrap(err, "from GetResourceList")
}
