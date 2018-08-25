// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе функции для отправки и чтения электронной почты.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/email"
	"github.com/pkg/errors"
)

// GetAddressBook возвращает список всех возможных адресатов.
func (s *Session) GetAddressBook() (*dt.AddressBook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var addressBook *dt.AddressBook
	switch s.Serv.Type {
	case cp.FirstType:
		addressBook, err = t01.GetAddressBook(&s.Session)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return addressBook, errors.Wrap(err, "from GetAddressBook")
}

// GetEmailsList возвращает список электронных писем на одной странице.
func (s *Session) GetEmailsList(nBoxID, startInd, pageSize, sequence string) (*dt.EmailsList, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var emailsList *dt.EmailsList
	switch s.Serv.Type {
	case cp.FirstType:
		emailsList, err = t01.GetEmailsList(&s.Session, nBoxID, startInd, pageSize, sequence)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return emailsList, errors.Wrap(err, "from GetEmailsList")
}

// GetEmailDescription возвращает подробности заданного электронного письма.
func (s *Session) GetEmailDescription(MID, MBID string) (*dt.EmailDescription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var emailDescription *dt.EmailDescription
	switch s.Serv.Type {
	case cp.FirstType:
		emailDescription, err = t01.GetEmailDescription(&s.Session, MID, MBID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return emailDescription, errors.Wrap(err, "from GetEmailDescription")
}

// CreateEmail создает сообщение и отправляет его адресатам.
func (s *Session) CreateEmail(userID, LBC, LCC, LTO, name, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.CreateEmail(&s.Session, userID, LBC, LCC, LTO, name, message)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from CreateEmail")
}

// DeleteEmails удаяет заданные сообщения.
func (s *Session) DeleteEmails(boxID string, emailIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.DeleteEmails(&s.Session, boxID, emailIDs)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from DeleteMessages")
}
