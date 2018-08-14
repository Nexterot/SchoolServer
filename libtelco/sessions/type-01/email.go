// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе функции для отправки и чтения электронной почты.
*/
package type01

import (
	"bytes"
	"fmt"
	"strconv"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions/session"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetEmailsList возвращает список электронных писем на одной странице с сервера первого типа.
func GetEmailsList(s *ss.Session, nBoxID, startInd, pageSize, sequence string) (*dt.EmailsList, error) {
	p := "http://"

	// 0-ой Get-запрос (не дублирующийся).
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER), ro)
		return true, err
	}
	_, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 GET")
	}

	// 1-ый POST-запрос.
	r1 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+
			fmt.Sprintf("/asp/ajax/GetMessagesAjax.asp?AT=%v&nBoxID=%v&jtStartIndex=%v&jtPageSize=%v&jtSorting=Sent%%20%v",
				s.AT, nBoxID, startInd, pageSize, sequence), ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		b, flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней список всех сообщений.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	var findMailLettersListTableNode func(*html.Node) *html.Node
	findMailLettersListTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "jtable" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findMailLettersListTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	formMailLettersList := func(node *html.Node) ([]dt.Email, error) {
		letters := make([]dt.Email, 0, 1)
		if node == nil {
			return letters, errors.New("Node is nil in func formEmailsList")
		}
		if node.FirstChild == nil {
			return letters, errors.New("Couldn't find list of letters in func formEmailsList")
		}
		infoNode := node.FirstChild
		if infoNode.NextSibling == nil {
			return letters, errors.New("Couldn't find list of letters in func formEmailsList")
		}
		infoNode = infoNode.NextSibling
		for letterNode := infoNode.FirstChild; letterNode != nil; letterNode = letterNode.NextSibling {
			if letterNode.FirstChild == nil {
				continue
			}
			c := letterNode.FirstChild
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling

			// Нашли письмо
			letter := dt.Email{}
			if c.FirstChild != nil {
				if len(c.FirstChild.Attr) != 0 {
					// Находим ID
					for _, a := range c.FirstChild.Attr {
						if a.Key == "href" {
							var start, end int
							for i := 0; i < len(a.Val); i++ {
								if a.Val[i:i+1] == "(" {
									start = i + 1
								}
								if a.Val[i:i+1] == "," {
									end = i
									break
								}
							}
							letter.ID, err = strconv.Atoi(a.Val[start:end])
							if err != nil {
								return letters, err
							}
						}
					}
				}
				if c.FirstChild.FirstChild != nil {
					// Находим автора письма
					letter.Author = c.FirstChild.FirstChild.Data
				}
			}
			if c.NextSibling == nil {
				letters = append(letters, letter)
				continue
			}
			c = c.NextSibling
			if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
				// Находим тему письма
				letter.Title = c.FirstChild.FirstChild.Data
			}
			if c.NextSibling == nil {
				letters = append(letters, letter)
				continue
			}
			c = c.NextSibling
			if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
				// Находим дату письма
				letter.Date = c.FirstChild.FirstChild.Data
			}

			letters = append(letters, letter)
		}

		return letters, nil
	}

	makeMailLettersList := func(node *html.Node) (*dt.EmailsList, error) {
		letters := &dt.EmailsList{}
		tableNode := findMailLettersListTableNode(node)
		letters.Letters, err = formMailLettersList(tableNode)

		return letters, err
	}

	return makeMailLettersList(parsedHTML)
}
