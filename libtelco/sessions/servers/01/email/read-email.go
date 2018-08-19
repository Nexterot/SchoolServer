package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetEmailsList возвращает список электронных писем на одной странице с сервера первого типа.
func GetEmailsList(s *dt.Session, nBoxID, startInd, pageSize, sequence string) (*dt.EmailsList, error) {
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
			Data:    map[string]string{},
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

	// Если мы дошли до этого места, то можно распарсить JSON,
	// находящийся в теле ответа, и найти в нем список всех сообщений.
	emailsList := &dt.EmailsList{}
	if err := json.Unmarshal(b, emailsList); err != nil {
		return nil, err
	}
	return emailsList, nil
}

// GetEmailDescription возвращает подробности заданного электронного письма с сервера первого типа.
func GetEmailDescription(s *dt.Session, MID, MBID string) (*dt.EmailDescription, error) {
	p := "http://"

	// 0-ой Get-запрос (не дублирующийся).
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{}
		r, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/readmessage.asp?at=%s&ver=%s&MID=%s&MBID=%s", s.AT, s.VER, MID, MBID), ro)
		if err != nil {
			return nil, false, err
		}
		return r.Bytes(), true, err
	}
	b, _, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 GET")
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней список всех сообщений.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	var findEmailDescriptionTableNode func(*html.Node) *html.Node
	findEmailDescriptionTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "form" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "name" && a.Val == "ReadMsg" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findEmailDescriptionTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит из переданной строки путь к файлу и его attachmentId
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
			return url, id, errors.New("Couldn't find url path or attachment id of file in task details")
		}

		return url, id, nil
	}

	formEmailDescription := func(node *html.Node) (*dt.EmailDescription, error) {
		description := dt.EmailDescription{}
		if node == nil {
			return &description, errors.New("Node is nil in func formMailLetterDescription")
		}
		if node.FirstChild == nil {
			return &description, errors.New("Couldn't find mail description in func formMailLetterDescription")
		}
		infoNode := node.FirstChild
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}
		if infoNode == nil {
			return &description, errors.New("Couldn't find mail description in func formMailLetterDescription")
		}
		if infoNode.NextSibling == nil {
			return &description, errors.New("Couldn't find mail description in func formMailLetterDescription")
		}
		infoNode = infoNode.NextSibling
		if infoNode.NextSibling == nil {
			return &description, errors.New("Couldn't find mail description in func formMailLetterDescription")
		}
		infoNode = infoNode.NextSibling

		// Ищем заголовки письма (кому было отправлено письмо и копию)
		c := infoNode.FirstChild
		for c != nil && c.Data != "div" {
			c = c.NextSibling
		}
		if c != nil {
			c = c.FirstChild
			for c != nil && c.Data != "div" {
				c = c.NextSibling
			}
			if c != nil {
				c = c.FirstChild
				flag := true
				for c != nil && flag {
					if c.Data == "div" {
						for _, a := range c.Attr {
							if a.Key == "id" && a.Val == "message_headers" {
								flag = false
								break
							}
						}
					}
					if flag {
						c = c.NextSibling
					}
				}
				if c != nil && c.FirstChild != nil {
					c = c.FirstChild
					for c != nil {
						if c.Data == "div" {
							break
						}
						c = c.NextSibling
					}
					if c != nil {
						c = c.FirstChild
						for c != nil && c.Data != "div" {
							c = c.NextSibling
						}
						if c.NextSibling != nil && c.NextSibling.Data == "div" {
							c = c.NextSibling
							// c2 - node, содержащий копии
							c2 := c.NextSibling

							// Ищем, кому было отправлено письмо
							c = c.FirstChild
							for c != nil && c.Data != "div" {
								c = c.NextSibling
							}
							if c != nil && c.FirstChild != nil {
								c = c.FirstChild
								for c != nil && c.Data != "input" {
									c = c.NextSibling
								}
								if c != nil {
									// Нашли, кому было послано письмо
									for _, a := range c.Attr {
										if a.Key == "value" && len(a.Val) != 0 {
											strs := strings.Split(a.Val, "; ")
											for i := 0; i < len(strs); i++ {
												mailUser := dt.EmailUser{}
												mailUser.Name = strs[i]
												description.To = append(description.To, mailUser)
											}
										}
									}
								}
							}

							// Ищем копии
							if c2 != nil {
								c2 = c2.FirstChild
								for c2 != nil && c2.Data != "div" {
									c2 = c2.NextSibling
								}
								if c2 != nil && c2.FirstChild != nil {
									c2 = c2.FirstChild
									for c2 != nil && c2.Data != "input" {
										c2 = c2.NextSibling
									}
									if c2 != nil {
										// Нашли копии
										for _, a := range c2.Attr {
											if a.Key == "value" && len(a.Val) != 0 {
												strs := strings.Split(a.Val, "; ")
												for i := 0; i < len(strs); i++ {
													mailUser := dt.EmailUser{}
													mailUser.Name = strs[i]
													description.Copy = append(description.Copy, mailUser)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// Есть ли в письме прикреплённые файлы
		hasFiles := false

		// Находим тело письма
		if infoNode.NextSibling != nil {
			infoNode = infoNode.NextSibling
			if infoNode.NextSibling == nil {
				return &description, errors.New("Couldn't find body of letter description")
			}
			infoNode = infoNode.NextSibling
			if infoNode.FirstChild != nil {
				infoNode = infoNode.FirstChild
				for infoNode != nil && infoNode.Data != "div" {
					infoNode = infoNode.NextSibling
				}
				if infoNode != nil && infoNode.FirstChild != nil {
					infoNode = infoNode.FirstChild
					if infoNode.NextSibling == nil {
						return &description, errors.New("Couldn't find body of letter description")
					}
					infoNode = infoNode.NextSibling
					if infoNode.FirstChild != nil {
						infoNode = infoNode.FirstChild
						for infoNode != nil && infoNode.Data != "div" {
							infoNode = infoNode.NextSibling
						}
						if infoNode != nil && infoNode.NextSibling != nil && infoNode.NextSibling.NextSibling != nil {
							infoNode = infoNode.NextSibling.NextSibling
							if infoNode.FirstChild != nil {
								infoNode = infoNode.FirstChild
								for infoNode != nil && infoNode.Data != "div" {
									infoNode = infoNode.NextSibling
								}
								if infoNode != nil && infoNode.FirstChild != nil {
									infoNode = infoNode.FirstChild
									for infoNode != nil && infoNode.Data != "div" {
										infoNode = infoNode.NextSibling
									}
									if infoNode != nil && infoNode.FirstChild != nil {
										infoNode = infoNode.FirstChild
										for infoNode != nil && infoNode.Data != "div" {
											infoNode = infoNode.NextSibling
										}
										if infoNode != nil && infoNode.NextSibling != nil {
											infoNode = infoNode.NextSibling
											if infoNode.FirstChild != nil {
												infoNode = infoNode.FirstChild
												for infoNode != nil && infoNode.Data != "label" {
													infoNode = infoNode.NextSibling
												}
												if infoNode.FirstChild != nil {
													if infoNode.FirstChild.Data == "Присоединенные файлы" {
														hasFiles = true
													}
												}

												if hasFiles {
													c2 := infoNode.Parent
													c2 = c2.NextSibling
													for c2 != nil && c2.Data != "div" {
														c2 = c2.NextSibling
													}

													for infoNode != nil && infoNode.Data != "div" {
														infoNode = infoNode.NextSibling
													}
													if infoNode != nil && infoNode.FirstChild != nil {
														infoNode = infoNode.FirstChild
														for infoNode != nil && infoNode.Data != "span" {
															infoNode = infoNode.NextSibling
														}
														for infoNode != nil {
															if infoNode.Data == "span" {
																if infoNode != nil && infoNode.FirstChild != nil {
																	for fileNode := infoNode.FirstChild; fileNode != nil; fileNode = fileNode.NextSibling {
																		if fileNode.FirstChild != nil {
																			file := dt.EmailFile{}
																			for _, a := range fileNode.Attr {
																				if a.Key == "href" {
																					file.Path, file.ID, err = findURLAndID(a.Val)
																					if err != nil {
																						return &description, err
																					}
																					break
																				}
																			}
																			file.FileName = fileNode.FirstChild.Data
																			description.Files = append(description.Files, file)
																		}
																	}
																}
															}
															infoNode = infoNode.NextSibling
														}
													}

													infoNode = c2.FirstChild
												}

												for infoNode != nil && infoNode.Data != "div" {
													infoNode = infoNode.NextSibling
												}
												// Нашли тело письма
												var str string
												for c = infoNode.FirstChild; c != nil; c = c.NextSibling {
													if c.FirstChild != nil {
														j := 0
														for c2 := c.FirstChild; c2 != nil; c2 = c2.NextSibling {
															if c2.Data == "br" {
																str += "\n"
															} else {
																str += c2.Data
															}
															j++
														}
														if j > 1 {
															str += "\n"
														}
													} else {
														if c.Data == "br" {
															str += "\n"
														} else {
															str += c.Data
														}
													}
												}
												description.Description = str
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return &description, nil
	}

	makeEmailDescription := func(node *html.Node) (*dt.EmailDescription, error) {
		tableNode := findEmailDescriptionTableNode(node)
		return formEmailDescription(tableNode)
	}

	return makeEmailDescription(parsedHTML)
}
