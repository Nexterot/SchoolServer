// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе функции для отправки и чтения электронной почты.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/type-01"
	"github.com/pkg/errors"
)

/*
Получение списка писем.
*/

// GetEmailsList возвращает список электронных писем на одной странице.
func (s *Session) GetEmailsList(nBoxID, startInd, pageSize, sequence string) (*dt.EmailsList, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	var emailsList *dt.EmailsList
	switch s.Base.Serv.Type {
	case cp.FirstType:
		emailsList, err = t01.GetEmailsList(s.Base, nBoxID, startInd, pageSize, sequence)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return emailsList, errors.Wrap(err, "from GetEmailsList")
}
