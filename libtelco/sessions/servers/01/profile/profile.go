// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package profile

import (
	"bytes"
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetProfile получает подробности профиля с сервера первого типа.
func GetProfile(s *dt.Session) (*dt.Profile, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"UID": "",
				"VER": s.VER,
				"at":  s.AT,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/MySettings/MySettings.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		if err != nil {
			return nil, false, err
		}
		return r.Bytes(), flag, nil
	}
	b, flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней мапу детей в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Находит node с данными профиля пользователя
	var findProfileTableNode func(*html.Node) *html.Node
	findProfileTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "col-xs-12 col-sm-12 col-md-10 col-lg-8 col-md-pull-2" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findProfileTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит node с UID пользователя
	var findUIDNode func(*html.Node) *html.Node
	findUIDNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "form" {
				for _, a := range node.Attr {
					if a.Key == "action" && (a.Val == "SaveParentSettings.asp" || a.Val == "SaveMySettings.asp") {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findUIDNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	findUID := func(node *html.Node) (string, error) {
		if node == nil {
			return "", errors.New("Node is nil")
		}
		if node.FirstChild == nil {
			return "", errors.New("Couldn't find UID")
		}
		for infoNode := node.FirstChild; infoNode != nil; infoNode = infoNode.NextSibling {
			if infoNode.Data == "input" {
				for _, a := range infoNode.Attr {
					if a.Key == "name" && a.Val == "UID" {
						for _, a2 := range infoNode.Attr {
							if a2.Key == "value" {
								return a2.Val, nil
							}
						}
					}
				}
			}
		}
		return "", errors.New("Couldn't find UID after searching all siblings")
	}

	// Формирует профиль
	formProfile := func(node *html.Node) (*dt.Profile, error) {
		profile := &dt.Profile{}
		if node == nil {
			return nil, errors.New("Node is nil")
		}
		if node.FirstChild == nil {
			return nil, errors.New("Couldn't find profile settings")
		}
		for infoNode := node.FirstChild; infoNode != nil; infoNode = infoNode.NextSibling {
			if infoNode.Data == "div" && infoNode.FirstChild != nil {
				var key, val string
				dataNode := infoNode.FirstChild
				for dataNode != nil && dataNode.Data != "label" {
					dataNode = dataNode.NextSibling
				}
				if dataNode == nil || dataNode.FirstChild == nil {
					continue
				}
				key = dataNode.FirstChild.Data
				if key == "Текущий учебный год" {
					for dataNode != nil && dataNode.Data != "div" {
						dataNode = dataNode.NextSibling
					}
					if dataNode == nil || dataNode.FirstChild == nil {
						continue
					}
					dataNode = dataNode.FirstChild
					for dataNode != nil && dataNode.Data != "select" {
						dataNode = dataNode.NextSibling
					}
					if dataNode == nil || dataNode.FirstChild == nil {
						continue
					}
					for dataNode = dataNode.FirstChild; dataNode != nil; dataNode = dataNode.NextSibling {
						if dataNode.Data == "option" {
							for _, a := range dataNode.Attr {
								if a.Key == "selected" {
									if dataNode.FirstChild != nil {
										profile.Schoolyear = dataNode.FirstChild.Data
									}
								}
							}
						}
					}
					continue
				} else {
					if key == "Роль в системе" {
						for dataNode != nil && dataNode.Data != "div" {
							dataNode = dataNode.NextSibling
						}
						if dataNode == nil || dataNode.FirstChild == nil {
							continue
						}
						roleNode := dataNode.FirstChild
						for roleNode != nil && roleNode.Data != "input" {
							roleNode = roleNode.NextSibling
						}
						if roleNode == nil {
							profile.Role = dataNode.FirstChild.Data
							continue
						} else {
							for _, a := range roleNode.Attr {
								if a.Key == "value" {
									profile.Role = a.Val
									continue
								}
							}
							continue
						}
					}
				}
				for dataNode != nil && dataNode.Data != "div" {
					dataNode = dataNode.NextSibling
				}
				if dataNode == nil || dataNode.FirstChild == nil {
					continue
				}
				dataNode = dataNode.FirstChild
				for dataNode != nil && dataNode.Data != "input" {
					dataNode = dataNode.NextSibling
				}
				if dataNode == nil {
					continue
				}
				for _, a := range dataNode.Attr {
					if a.Key == "value" {
						val = a.Val
					}
				}
				switch key {
				case "Фамилия":
					profile.Surname = val
				case "Имя":
					profile.Name = val
				case "Имя пользователя":
					profile.Username = val
				}
			}
		}

		return profile, nil
	}

	// Создаёт профиль пользователя
	makeProfile := func(node *html.Node) (*dt.Profile, error) {
		tableNode := findProfileTableNode(node)
		profile, err := formProfile(tableNode)
		if err != nil {
			return profile, err
		}
		profile.UID, err = findUID(findUIDNode(node))
		return profile, err
	}

	return makeProfile(parsedHTML)
}
