package forum

import (
	"bytes"
	"fmt"
	"strconv"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetForumThemesList возвращает список тем форума c сервера первого типа.
func GetForumThemesList(s *dt.Session, page string) (*dt.ForumThemesList, error) {
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

	// 1-ый POST-запрос.
	r1 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"BACK":      "",
				"DELARR":    "",
				"LoginType": "0",
				"PAGE":      page,
				"PAGESIZE":  "25",
				"VER":       s.VER,
				"\"":        "",
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Forum/Forum.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+
			fmt.Sprintf("/asp/Forum/Forum.asp?PAGE=%s&PAGESIZE=25", page), ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
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
	// находящуюся в теле ответа, и найти в ней список всех тем форума.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
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

// GetForumThemeMessages возвращает список всех сообщений одной темы форума с сервера первого типа.
func GetForumThemeMessages(s *dt.Session, TID, page, pageSize string) (*dt.ForumThemeMessages, error) {
	p := "http://"

	// 0-ой POST-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"BACK":      "",
				"DELARR":    "",
				"LoginType": "0",
				"PAGE":      page,
				"PAGESIZE":  pageSize,
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Forum/Forum.asp?PAGE=1&PAGESIZE=25",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+
			fmt.Sprintf("/asp/Forum/ShowThread.asp?TID=%s", TID), ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
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
	// находящуюся в теле ответа, и найти в ней список всех сообщений одной темы форума.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	var findForumPostMessagesTableNode func(*html.Node) *html.Node
	findForumPostMessagesTableNode = func(node *html.Node) *html.Node {
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
			n := findForumPostMessagesTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	formForumPostMessages := func(node *html.Node) ([]dt.ForumThemeMessage, error) {
		messages := make([]dt.ForumThemeMessage, 0, 1)
		if node == nil {
			return messages, errors.New("Node is nil in func formForumPostMessages")
		}
		if node.FirstChild == nil {
			return messages, errors.New("Couldn't find messages in func formForumPostMessages")
		}
		infoNode := node.FirstChild
		if infoNode.NextSibling == nil {
			return messages, errors.New("Couldn't find messages in func formForumPostMessages")
		}
		infoNode = infoNode.NextSibling
		if infoNode.FirstChild == nil {
			return messages, errors.New("Couldn't find messages in func formForumPostMessages")
		}
		for messageNode := infoNode.FirstChild.NextSibling; messageNode != nil; messageNode = messageNode.NextSibling {
			if messageNode.FirstChild == nil {
				continue
			}
			message := dt.ForumThemeMessage{}
			c := messageNode.FirstChild
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling

			if c.FirstChild != nil {
				authorNode := c.FirstChild
				if authorNode.FirstChild != nil {
					message.Author = authorNode.FirstChild.Data
				}
				if authorNode.NextSibling != nil && authorNode.NextSibling.NextSibling != nil && authorNode.NextSibling.NextSibling.FirstChild != nil {
					message.Role = authorNode.NextSibling.NextSibling.FirstChild.Data
				}
			}
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling
			if c.FirstChild == nil {
				continue
			}
			c = c.FirstChild

			if c.FirstChild != nil {
				for i := 0; i < len(c.FirstChild.Data); i++ {
					if c.FirstChild.Data[i:i+1] == ":" {
						message.Date = c.FirstChild.Data[i+2:]
						break
					}
				}
			}

			if c.NextSibling == nil {
				continue
			}

			c = c.NextSibling
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling
			if c.NextSibling == nil {
				continue
			}
			c = c.NextSibling
			if c.FirstChild == nil {
				continue
			}
			c = c.FirstChild
			var messageString string
			for c != nil {
				if c.Data != "br" {
					messageString += c.Data
				} else {
					messageString += "\n"
				}
				c = c.NextSibling
			}
			message.Message = messageString

			messages = append(messages, message)
		}

		return messages, err
	}

	makeForumPostMessages := func(node *html.Node) (*dt.ForumThemeMessages, error) {
		messages := dt.ForumThemeMessages{}
		tableNode := findForumPostMessagesTableNode(node)
		messages.Messages, err = formForumPostMessages(tableNode)

		return &messages, err
	}

	return makeForumPostMessages(parsedHTML)
}
