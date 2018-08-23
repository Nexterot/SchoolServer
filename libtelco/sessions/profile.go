// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/profile"
	"github.com/pkg/errors"
)

/*
Получение подробностей профиля.
*/

// GetProfile получает подробности профиля.
func (s *Session) GetProfile() (*dt.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var profile *dt.Profile
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		profile, err = t01.GetProfile(&s.Session)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return profile, errors.Wrap(err, "from GetProfile")
}
