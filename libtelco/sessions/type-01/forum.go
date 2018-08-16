// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе функции для отправки и чтения сообщений на форуме.
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

// GetForumThemesList возвращает список тем форума c сервера первого типа.
func GetForumThemesList(s *ss.Session) (*dt.ForumThemesList, error) {
	p := "http://"

	// 0-ой Get-запрос (не дублирующийся).
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Forum/Forum.asp?AT=%s&VER=%s", s.AT, s.VER), ro)
		return true, err
	}
	_, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 GET")
	}

	// Если мы дошли до этого места, то можно распарсить JSON,
	// находящийся в теле ответа, и найти в нем список всех сообщений.
	parsedHTML, err := html.Parse(bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	var findForumThemesTableNode func(*html.Node) *html.Node
	findForumThemesTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "table table-bordered" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findForumThemesTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	formForumThemesList := func(node *html.Node) ([]dt.ForumTheme, error) {
		themes := make([]dt.ForumTheme, 0)
		if node == nil {
			return themes, errors.New("Node is nil in func formForumThemesList")
		}
		if node.FirstChild == nil {
			return themes, errors.New("Couldn't find list of themes in func formForumThemesList")
		}
		infoNode := node.FirstChild
		if infoNode.NextSibling == nil {
			return themes, errors.New("Couldn't find list of themes in func formForumThemesList")
		}
		infoNode = infoNode.NextSibling
		if infoNode.FirstChild == nil {
			return themes, errors.New("Couldn't find list of themes in func formForumThemesList")
		}
		for themeNode := infoNode.FirstChild.NextSibling; themeNode != nil; themeNode = themeNode.NextSibling {
			if themeNode.FirstChild == nil {
				continue
			}
			// Формируем найденную тему и добавляем её в список тем
			theme := dt.ForumTheme{}
			c := themeNode.FirstChild
			if c.FirstChild != nil && c.FirstChild.FirstChild != nil {
				theme.Title = c.FirstChild.FirstChild.Data

				// Находим ID темы в c.FirstChild.Data
				var str, stringID string
				for _, a := range c.FirstChild.Attr {
					if a.Key == "href" {
						str = a.Val
					}
				}

				for i := 0; i < len(str); i++ {
					if str[i:i+1] == "(" {
						stringID = str[i+1 : len(str)-1]
						break
					}
				}
				theme.ID, err = strconv.Atoi(stringID)
				if err != nil {
					return themes, err
				}
			}

			if c.NextSibling == nil {
				themes = append(themes, theme)
				continue
			}
			c = c.NextSibling
			if c.FirstChild != nil {
				theme.Creator = c.FirstChild.Data
			}

			if c.NextSibling == nil {
				themes = append(themes, theme)
				continue
			}
			c = c.NextSibling
			if c.NextSibling == nil {
				themes = append(themes, theme)
				continue
			}
			c = c.NextSibling
			if c.FirstChild != nil {
				theme.Answers, err = strconv.Atoi(c.FirstChild.Data)
				if err != nil {
					return themes, err
				}
			}

			if c.NextSibling == nil {
				themes = append(themes, theme)
				continue
			}
			c = c.NextSibling
			if c.FirstChild != nil {
				c = c.FirstChild
				if c.FirstChild != nil {
					theme.Date = c.FirstChild.Data
				}
				if c.NextSibling == nil {
					themes = append(themes, theme)
					continue
				}
				c = c.NextSibling
				if c.NextSibling == nil {
					themes = append(themes, theme)
					continue
				}
				c = c.NextSibling
				theme.LastAuthor = c.Data
			}

			themes = append(themes, theme)
		}

		return themes, err
	}

	makeForumThemesList := func(node *html.Node) (*dt.ForumThemesList, error) {
		themes := &dt.ForumThemesList{}
		tableNode := findForumThemesTableNode(node)
		themes.Posts, err = formForumThemesList(tableNode)

		return themes, err
	}

	return makeForumThemesList(parsedHTML)
}
