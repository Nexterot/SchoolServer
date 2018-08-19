// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе получение ресурсов.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/resources"

	"github.com/pkg/errors"
)

// GetResourcesList возвращает список всех ресурсов.
func (s *Session) GetResourcesList() (*dt.Resources, error) {
	s.MU.Lock()
	defer s.MU.Unlock()
	var err error
	var resources *dt.Resources
	switch s.Serv.Type {
	case cp.FirstType:
		resources, err = t01.GetResourcesList(&s.Session)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return resources, errors.Wrap(err, "from GetResourceList")
}
