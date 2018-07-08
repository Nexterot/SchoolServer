// Copyright (C) 2018 Mikhail Masyagin

/*
Package parser - данный файл содержит в себе сессию парсера на нужном школьном сайте.
*/
package parser

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// session struct содержит в себе описание сессии к одному из школьных серверов.
type session struct {
	// Общая структура.
	sess     *gr.Session
	serv     *cp.SchoolServer
	ready    bool
	reload   chan struct{}
	ignTimer chan struct{}
	logger   *log.Logger
	mu       sync.Mutex
	// Для серверов первого типа.
	at  string
	ver string
}

// newSession создает новую сессию на базе информации о школьном сервере,
// к которому предстоит подключиться.
func newSession(server *cp.SchoolServer, logger *log.Logger) *session {
	return &session{
		sess:     nil,
		serv:     server,
		ready:    false,
		reload:   make(chan struct{}),
		ignTimer: make(chan struct{}),
		logger:   logger,
		mu:       sync.Mutex{},
	}
}

// timer раз в s.serv.Time секунд отправляет запрос на перезагрузку go-рутины.
func (s *session) timer() {
	time.Sleep(time.Duration(s.serv.Time) * time.Second)
	s.reload <- struct{}{}
	select {
	case <-s.ignTimer:
		return
	default:
		s.reload <- struct{}{}
	}
}

// startSession подключается к серверу и держит с ним соединение всё отведенное время.
// Как только время заканчивается (например, на 62.117.74.43 стоит убогое ограничение в 45 минут,
// мы заново коннектимся).
func (s *session) startSession() {
	flag := false
	for {
		// Подключаемся к серверу.
		s.mu.Lock()
		s.ready = false
		if err := s.login(); err != nil {
			s.logger.Error("Error occured, while connecting to server",
				"Type", s.serv.Type,
				"Login", s.serv.Login,
				"Password", s.serv.Password,
				"error", err)
		} else {
			s.ready = true
			if !flag {
				s.logger.Info("New session was successfully created")
				flag = true
			} else {
				s.logger.Info("Session was successfully reloaded")
			}
			go s.timer()
		}
		s.mu.Unlock()
		// Ожидание требования перезагрузки.
		<-s.reload
	}
}

// login логинится к серверу и создает очередную сессию.
func (s *session) login() error {
	var err error
	switch s.serv.Type {
	case cp.FirstType:
		err = s.firstTypeLogin()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.serv.Type)
	}
	return err
}

// firstTypeLogin логинится к серверу первого типа и создает очередную сессию.
func (s *session) firstTypeLogin() error {
	// Создание сессии.
	s.sess = gr.NewSession(nil)
	p := "http://"
	// Полчение формы авторизации.
	// 0-ой Get-запрос.
	response0, err := s.sess.Get(p+s.serv.Link+"/asp/ajax/getloginviewdata.asp", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = response0.Close()
	}()
	// 1-ый Post-запрос.
	response1, err := s.sess.Post(p+s.serv.Link+"/webapi/auth/getdata", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = response1.Close()
	}()
	type FirstAnswer struct {
		Lt   string `json:"lt"`
		Ver  string `json:"ver"`
		Salt string `json:"salt"`
	}
	fa := &FirstAnswer{}
	if err = response1.JSON(fa); err != nil {
		return err
	}
	// 2-ой Post-запрос.
	pw := s.serv.Password
	hasher := md5.New()
	if _, err = hasher.Write([]byte(pw)); err != nil {
		return err
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	hasher = md5.New()
	if _, err = hasher.Write([]byte(fa.Salt + pw)); err != nil {
		return err
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	requestOption := &gr.RequestOptions{
		Data: map[string]string{
			"CID":       "2",
			"CN":        "1",
			"LoginType": "1",
			"PID":       "-1",
			"PW":        pw[:len(s.serv.Password)],
			"SCID":      "2",
			"SFT":       "2",
			"SID":       "77",
			"UN":        s.serv.Login,
			"lt":        fa.Lt,
			"pw2":       pw,
			"ver":       fa.Ver,
		},
		Headers: map[string]string{
			"Referer": p + s.serv.Link + "/",
		},
	}
	response2, err := s.sess.Post(p+s.serv.Link+"/asp/postlogin.asp", requestOption)
	if err != nil {
		return err
	}
	defer func() {
		_ = response2.Close()
	}()
	fmt.Println(string(response2.Bytes()))
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней "AT".
	parsedHTML, err := html.Parse(bytes.NewReader(response2.Bytes()))
	if err != nil {
		return err
	}
	var f func(*html.Node, string) string
	f = func(node *html.Node, reqAttr string) string {
		if node.Type == html.ElementNode {
			for i := 0; i < len(node.Attr)-1; i++ {
				tmp0 := node.Attr[i]
				tmp1 := node.Attr[i+1]
				if (tmp0.Key == "name") && (tmp0.Val == reqAttr) && (tmp1.Key == "value") {
					return tmp1.Val
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if str := f(c, reqAttr); str != "" {
				return str
			}
		}
		return ""
	}

	s.at = f(parsedHTML, "AT")
	s.ver = f(parsedHTML, "VER")
	if (s.at == "") || (s.ver == "") {
		err = fmt.Errorf("Problems on school server: %s", s.serv.Link)
		return err
	}
	return nil
}
