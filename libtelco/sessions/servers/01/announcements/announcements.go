// Copyright (C) 2018 Mikhail Masyagin

package announcements

import (
	"bytes"
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func GetAnnouncements(s *dt.Session) (*dt.Posts, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"VER": s.VER,
				"at":  s.AT,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Announce/ViewAnnouncements.asp", ro)
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

	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Находит node с данными профиля пользователя
	var findPostsTableNode func(*html.Node) *html.Node
	findPostsTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "adver-container" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findPostsTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит из переданной строки путь к файлу и его attachmentID
	findURLAndID := func(str string) (string, string, error) {
		var url, id string
		var i, j int
		for i = 0; i < len(str); i++ {
			if str[i:i+1] == "(" {
				break
			}
		}
		for j = i; j < len(str); j++ {
			if str[j:j+1] == "," {
				url = str[i+2 : j-1]
				break
			}
		}
		for i = j + 1; i < len(str); i++ {
			if str[i:i+1] == ")" {
				id = str[j+2 : i]
			}
		}
		if url == "" || id == "" {
			return url, id, errors.New("couldn't find url path or attachment id")
		}

		return url, id, nil
	}

	// Формирует профиль
	formPosts := func(node *html.Node) (*dt.Posts, error) {
		posts := dt.Posts{}
		if node == nil {
			return &posts, errors.New("node is nil")
		}
		if node.FirstChild == nil {
			return &posts, errors.New("couldn't find posts")
		}
		for postNode := node.FirstChild; postNode != nil; postNode = postNode.NextSibling {
			if postNode.Data != "div" || postNode.FirstChild == nil {
				continue
			}
			infoNode := postNode.FirstChild
			for infoNode != nil && infoNode.Data != "div" {
				infoNode = infoNode.NextSibling
			}
			if infoNode == nil {
				continue
			}
			post := dt.Post{}
			infoNode = infoNode.FirstChild
			for infoNode != nil && infoNode.Data != "h3" {
				infoNode = infoNode.NextSibling
			}
			if infoNode != nil && infoNode.FirstChild != nil {
				themeNode := infoNode.FirstChild
				for themeNode != nil && themeNode.Data != "span" {
					themeNode = themeNode.NextSibling
				}
				if themeNode != nil {
					themeNode = themeNode.NextSibling
					if themeNode != nil {
						post.Title = themeNode.Data
					}
				}
			}
			for infoNode != nil && infoNode.Data != "div" {
				infoNode = infoNode.NextSibling
			}
			if infoNode != nil && infoNode.FirstChild != nil {
				authorNode := infoNode.FirstChild
				for authorNode != nil && authorNode.Data != "a" {
					authorNode = authorNode.NextSibling
				}
				if authorNode != nil && authorNode.FirstChild != nil {
					post.Author = authorNode.FirstChild.Data
				}
			}
			if infoNode == nil {
				continue
			}
			infoNode = infoNode.NextSibling
			for infoNode != nil && infoNode.Data != "div" {
				infoNode = infoNode.NextSibling
			}
			if infoNode != nil && infoNode.FirstChild != nil {
				dateNode := infoNode.FirstChild
				for dateNode != nil && dateNode.Data != "span" {
					dateNode = dateNode.NextSibling
				}
				if dateNode != nil && dateNode.FirstChild != nil {
					post.Date = dateNode.FirstChild.Data
				}
			}
			if infoNode == nil {
				continue
			}
			infoNode = infoNode.NextSibling
			for infoNode != nil && infoNode.Data != "div" {
				infoNode = infoNode.NextSibling
			}
			if infoNode != nil && infoNode.FirstChild != nil {
				var message string
				for messageNode := infoNode.FirstChild; messageNode != nil; messageNode = messageNode.NextSibling {
					if messageNode.Data == "br" {
						message += "\n"
					} else {
						if messageNode.Data == "div" {
							if messageNode.FirstChild != nil {
								if messageNode.Data == "a" {
									// Нашли ссылку на сайт
									for _, a := range messageNode.Attr {
										if a.Key == "href" {
											message += a.Val
											break
										}
									}
								} else {
									// Нашли прикреплённый файл
									fileNode := messageNode.FirstChild
									for fileNode != nil && fileNode.Data != "div" {
										fileNode = fileNode.NextSibling
									}
									if fileNode != nil && fileNode.FirstChild != nil {
										fileNode = fileNode.FirstChild
										for fileNode != nil && fileNode.Data != "span" {
											fileNode = fileNode.NextSibling
										}
										if fileNode != nil && fileNode.FirstChild != nil {
											fileNode = fileNode.FirstChild
											for fileNode != nil && fileNode.Data != "a" {
												fileNode = fileNode.NextSibling
											}
											if fileNode != nil && fileNode.FirstChild != nil {
												post.FileName = fileNode.FirstChild.Data
												for _, a := range fileNode.Attr {
													if a.Key == "href" {
														path, ID, err := findURLAndID(a.Val)
														if err != nil {
															return &posts, err
														}
														post.FileLink = path
														post.FileID = ID
													}
												}
											}
										}
									}
								}
							}
						} else {
							message += messageNode.Data
						}
					}
				}
				// Убираем лишние пробелы в начале и конце объявления
				if len(message) > 1 {
					if message[0] == 10 && message[1] == 9 {
						var i int
						for i = 0; i < len(message); i++ {
							if message[i] != 10 && message[i] != 9 {
								break
							}
						}
						message = message[i:]
					}
				}
				if len(message) > 1 {
					if message[len(message)-1] == 9 && message[len(message)-2] == 9 {
						var i int
						for i = len(message) - 1; i >= 0; i-- {
							if message[i] != 10 && message[i] != 9 {
								break
							}
						}
						message = message[:i+1]
					}
				}
				post.Message = message
			}
			posts.Posts = append(posts.Posts, post)
		}

		return &posts, nil
	}

	// Создаёт профиль пользователя
	makePosts := func(node *html.Node) (*dt.Posts, error) {
		tableNode := findPostsTableNode(node)
		return formPosts(tableNode)
	}

	return makePosts(parsedHTML)
}
