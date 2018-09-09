// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package email

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

// GetAddressBook возвращает список всех возможных адресатов с сервера первого типа.
func GetAddressBook(s *dt.Session) (*dt.AddressBook, error) {
	p := "http://"

	// 0-ой Get-запрос (не дублирующийся).
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   "http://62.117.74.43/",
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER), ro)
		return true, err
	}
	_, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 GET")
	}

	// 1-ый POST-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"Referer":          p + s.Serv.Link + fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+
			fmt.Sprintf("/asp/ajax/GetMessagesAjax.asp?AT=%v&nBoxID=1&jtStartIndex=0&jtPageSize=10&jtSorting=Sent%%20DESC", s.AT), ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
	}
	flag, err := r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}

	// 2-ой Get-запрос (не дублирующийся).
	r2 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r2()
	if err != nil {
		return nil, errors.Wrap(err, "2 GET")
	}

	// 3-ий Get-запрос (не дублирующийся).
	r3 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
				"Referer":          p + s.Serv.Link + fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+"/vendor/pages/css/print-tables.min.css", ro)
		return true, err
	}
	_, err = r3()
	if err != nil {
		return nil, errors.Wrap(err, "3 GET")
	}

	// 4-ый Get-запрос (не дублирующийся).
	r4 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r4()
	if err != nil {
		return nil, errors.Wrap(err, "4 GET")
	}

	// 5-ый Get-запрос (не дублирующийся).
	r5 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkbottom.asp?AT=%v&VER=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r5()
	if err != nil {
		return nil, errors.Wrap(err, "5 GET")
	}

	// 6-ой Get-запрос (не дублирующийся).
	r6 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkright.asp?AT=%v&F=COMPOSE&FN=ATO&FA=LTO&VER=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r6()
	if err != nil {
		return nil, errors.Wrap(err, "6 GET")
	}

	// 7-ой Get-запрос (не дублирующийся).
	r7 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		r, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkleft.asp?AT=%v&VER=%v", s.AT, s.VER), ro)
		if err != nil {
			return nil, false, err
		}
		return r.Bytes(), true, err
	}
	b, _, err := r7()
	if err != nil {
		return nil, errors.Wrap(err, "7 GET")
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней список всех групп рассылки.
	addressBook := dt.NewAddressBook()

	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	var findAddressBookTableNode func(*html.Node) *html.Node
	findAddressBookTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "container-fluid" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAddressBookTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	formAddressBookInfo := func(node *html.Node) ([]dt.AddressBookGroup, error) {
		if node == nil {
			return nil, errors.New("Node is nil in func formAddressBookInfo")
		}
		if node.FirstChild == nil {
			return nil, errors.New("Couldn't find address book info in func formAddressBookInfo")
		}
		infoNode := node.FirstChild
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}

		// Определяем, что записано в html-коде
		groups := make([]dt.AddressBookGroup, 0, 5)
		if infoNode != nil && infoNode.FirstChild != nil {
			groupNode := infoNode.FirstChild
			for groupNode != nil && groupNode.Data != "div" {
				groupNode = groupNode.NextSibling
			}
			if groupNode != nil && groupNode.FirstChild != nil {
				groupNode = groupNode.FirstChild
				var groupNameNode *html.Node
				for groupNode != nil {
					if groupNode.FirstChild != nil {
						groupNameNode = groupNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "label" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							if groupNameNode.FirstChild.Data == "Группа" {
								break
							}
						}

					}
					groupNode = groupNode.NextSibling
				}

				if groupNode != nil {
					for groupNameNode != nil && groupNameNode.Data != "div" {
						groupNameNode = groupNameNode.NextSibling
					}
					if groupNameNode.FirstChild != nil {
						groupNameNode = groupNameNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "select" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							groupNameNode = groupNameNode.FirstChild
							for groupNameNode != nil {
								if groupNameNode.FirstChild != nil {
									group := dt.AddressBookGroup{}
									group.Title = groupNameNode.FirstChild.Data
									for _, a := range groupNameNode.Attr {
										if a.Key == "value" {
											group.Value = a.Val
											break
										}
									}

									groups = append(groups, group)
								}
								groupNameNode = groupNameNode.NextSibling
							}
						}
					}
				}
			}
		}

		return groups, nil
	}

	// Возвращает группу или класс в адресной книге в зависимости от того, что есть в переданном html-коде
	parseAddressBookTable := func(node *html.Node) ([]dt.AddressBookGroup, error) {
		tableNode := findAddressBookTableNode(node)
		return formAddressBookInfo(tableNode)
	}

	groups, err := parseAddressBookTable(parsedHTML)
	if err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно начать поссылать запросы по группам (по всем, кроме класса).
	classIndex := -1
	for i := 0; i < len(groups); i++ {
		if groups[i].Title == "классы" {
			classIndex = i
			continue
		}
		if err := getGroupMembers(s, &groups[i]); err != nil {
			return nil, errors.Wrapf(err, "from getGroupMembers: can't get members of group %V(%V)", groups[i].Title, groups[i].Value)
		}
	}
	addressBook.Groups = groups

	if classIndex == -1 {
		return addressBook, nil
	}

	// Если мы дошли до этого места, то можно начать поссылать запросы по классам.
	// 8-ой POST-запрос.
	r8 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Messages/addrbkleft.asp",
			},
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"CLASSES":   "",
				"FL":        groups[classIndex].Value,
				"LoginType": "0",
				"OrgType":   "0",
				"VER":       s.VER,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Messages/addrbkleft.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err = r8()
	if err != nil {
		return nil, errors.Wrap(err, "8 POST")
	}
	if !flag {
		b, flag, err = r8()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 8 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 8 POST")
		}
	}

	parsedHTML, err = html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	findAddressBookTableNode1 := func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "container-fluid" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAddressBookTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Определяет, что записано в переданном html-коде (группа или класс), формирует найденную структуру и возвращает её
	formAddressBookInfo1 := func(node *html.Node) ([]dt.AddressBookClass, error) {
		if node == nil {
			return nil, errors.New("Node is nil in func formAddressBookInfo")
		}
		if node.FirstChild == nil {
			return nil, errors.New("Couldn't find address book info in func formAddressBookInfo")
		}
		infoNode := node.FirstChild
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}

		// Определяем, что записано в html-коде
		var groupName string
		if infoNode != nil && infoNode.FirstChild != nil {
			groupNode := infoNode.FirstChild
			for groupNode != nil && groupNode.Data != "div" {
				groupNode = groupNode.NextSibling
			}
			if groupNode != nil && groupNode.FirstChild != nil {
				groupNode = groupNode.FirstChild
				var groupNameNode *html.Node
				for groupNode != nil {
					if groupNode.FirstChild != nil {
						groupNameNode = groupNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "label" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							if groupNameNode.FirstChild.Data == "Группа" {
								break
							}
						}

					}
					groupNode = groupNode.NextSibling
				}

				if groupNode != nil {
					for groupNameNode != nil && groupNameNode.Data != "div" {
						groupNameNode = groupNameNode.NextSibling
					}
					if groupNameNode.FirstChild != nil {
						groupNameNode = groupNameNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "select" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							groupNameNode = groupNameNode.FirstChild
							found := false
							for groupNameNode != nil {
								if found {
									break
								}
								for _, a := range groupNameNode.Attr {
									if a.Key == "selected" {
										if groupNameNode.FirstChild != nil {
											groupName = groupNameNode.FirstChild.Data
										}
										found = true
										break
									}
								}
								groupNameNode = groupNameNode.NextSibling
							}
						}
					}
				}
			}
		}

		if groupName == "классы" {
			classes := make([]dt.AddressBookClass, 0, 10)
			if infoNode.NextSibling != nil {
				infoNode = infoNode.NextSibling
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
						if infoNode != nil && infoNode.FirstChild != nil {
							classNode := infoNode.FirstChild
							for classNode != nil && classNode.Data != "div" {
								classNode = classNode.NextSibling
							}
							if classNode != nil && classNode.FirstChild != nil {
								classNode = classNode.FirstChild
								for classNode != nil && classNode.Data != "select" {
									classNode = classNode.NextSibling
								}
								if classNode != nil && classNode.FirstChild != nil {
									for classNode = classNode.FirstChild; classNode != nil; classNode = classNode.NextSibling {
										if classNode.FirstChild != nil {
											class := dt.AddressBookClass{}
											class.ClassName = classNode.FirstChild.Data
											for _, a2 := range classNode.Attr {
												if a2.Key == "value" {
													class.ID, err = strconv.Atoi(a2.Val)
												}
											}

											classes = append(classes, class)
										}
									}
									return classes, nil
								}
							}
						}
					}
				}
			}
		}

		return nil, nil
	}

	// Возвращает группу или класс в адресной книге в зависимости от того, что есть в переданном html-коде
	parseAddressBookTable1 := func(node *html.Node) ([]dt.AddressBookClass, error) {
		tableNode := findAddressBookTableNode1(node)
		return formAddressBookInfo1(tableNode)
	}

	classes, err := parseAddressBookTable1(parsedHTML)
	if err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно начать поссылать запросы по классам.
	for i := 0; i < len(classes); i++ {
		if err := getClassMembers(s, &classes[i]); err != nil {
			return nil, errors.Wrapf(err, "from getClassMembers: can't get members of class %V(%V)", classes[i].ClassName, classes[i].ID)
		}
	}

	addressBook.Classes = classes

	for i := 0; i < len(addressBook.Groups); i++ {
		if addressBook.Groups[i].Title == "классы" {
			addressBook.Groups = append(addressBook.Groups[:i], addressBook.Groups[i+1:]...)
		}
	}

	return addressBook, nil
}

func getGroupMembers(s *dt.Session, group *dt.AddressBookGroup) error {
	p := "http://"

	// 0-ой POST-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Messages/addrbkleft.asp",
			},
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"FL":        group.Value,
				"LoginType": "0",
				"OrgType":   "0",
				"VER":       s.VER,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Messages/addrbkleft.asp", ro)
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
		return errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней членов группы.
	g, _, err := parseGroupOrClass(b)
	if err != nil {
		return err
	}
	group.Users = g.Users
	return nil
}

func getClassMembers(s *dt.Session, class *dt.AddressBookClass) error {
	p := "http://"

	// 0-ой POST-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Messages/addrbkleft.asp",
			},
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"CLASSES":   strconv.Itoa(class.ID),
				"FL":        "C",
				"LoginType": "0",
				"OrgType":   "0",
				"VER":       s.VER,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Messages/addrbkleft.asp", ro)
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
		return errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней членов группы.
	_, c, err := parseGroupOrClass(b)
	if err != nil {
		return err
	}
	class.Users = c.Users
	return nil
}

func parseGroupOrClass(b []byte) (*dt.AddressBookGroup, *dt.AddressBookClass, error) {
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}

	var findAddressBookTableNode func(*html.Node) *html.Node
	findAddressBookTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" && len(node.Attr) != 0 {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "container-fluid" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAddressBookTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Определяет, что записано в переданном html-коде (группа или класс), формирует найденную структуру и возвращает её
	formAddressBookInfo := func(node *html.Node) (*dt.AddressBookGroup, *dt.AddressBookClass, error) {
		if node == nil {
			return nil, nil, errors.New("Node is nil in func formAddressBookInfo")
		}
		if node.FirstChild == nil {
			return nil, nil, errors.New("Couldn't find address book info in func formAddressBookInfo")
		}
		infoNode := node.FirstChild
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}

		// Определяем, что записано в html-коде
		var groupName string
		if infoNode != nil && infoNode.FirstChild != nil {
			groupNode := infoNode.FirstChild
			for groupNode != nil && groupNode.Data != "div" {
				groupNode = groupNode.NextSibling
			}
			if groupNode != nil && groupNode.FirstChild != nil {
				groupNode = groupNode.FirstChild
				var groupNameNode *html.Node
				for groupNode != nil {
					if groupNode.FirstChild != nil {
						groupNameNode = groupNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "label" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							if groupNameNode.FirstChild.Data == "Группа" {
								break
							}
						}

					}
					groupNode = groupNode.NextSibling
				}

				if groupNode != nil {
					for groupNameNode != nil && groupNameNode.Data != "div" {
						groupNameNode = groupNameNode.NextSibling
					}
					if groupNameNode.FirstChild != nil {
						groupNameNode = groupNameNode.FirstChild
						for groupNameNode != nil && groupNameNode.Data != "select" {
							groupNameNode = groupNameNode.NextSibling
						}
						if groupNameNode != nil && groupNameNode.FirstChild != nil {
							groupNameNode = groupNameNode.FirstChild
							found := false
							for groupNameNode != nil {
								if found {
									break
								}
								for _, a := range groupNameNode.Attr {
									if a.Key == "selected" {
										if groupNameNode.FirstChild != nil {
											groupName = groupNameNode.FirstChild.Data
										}
										found = true
										break
									}
								}
								groupNameNode = groupNameNode.NextSibling
							}
						}
					}
				}
			}
		}

		if groupName == "классы" {
			// В html-коде записан класс, составляем структуру AddressBookClass
			class := dt.AddressBookClass{}
			if infoNode.NextSibling != nil {
				infoNode = infoNode.NextSibling
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
						if infoNode != nil && infoNode.FirstChild != nil {
							classNode := infoNode.FirstChild
							for classNode != nil && classNode.Data != "div" {
								classNode = classNode.NextSibling
							}
							if classNode != nil && classNode.FirstChild != nil {
								classNode = classNode.FirstChild
								for classNode != nil && classNode.Data != "select" {
									classNode = classNode.NextSibling
								}
								if classNode != nil && classNode.FirstChild != nil {
									found := false
									for classNode = classNode.FirstChild; classNode != nil; classNode = classNode.NextSibling {
										if found {
											break
										}
										for _, a := range classNode.Attr {
											if a.Key == "selected" {
												if classNode.FirstChild != nil {
													class.ClassName = classNode.FirstChild.Data
												}
												for _, a2 := range classNode.Attr {
													if a2.Key == "value" {
														class.ID, err = strconv.Atoi(a2.Val)
													}
												}
												found = true
												break
											}
										}
									}
								}
							}
						}
						for infoNode != nil && infoNode.Data != "table" {
							infoNode = infoNode.NextSibling
						}
						if infoNode != nil && infoNode.FirstChild != nil {
							infoNode = infoNode.FirstChild
							for infoNode != nil && infoNode.Data != "tbody" {
								infoNode = infoNode.NextSibling
							}
							if infoNode != nil && infoNode.FirstChild != nil {
								for infoNode = infoNode.FirstChild; infoNode != nil; infoNode = infoNode.NextSibling {
									user := dt.AddressBookClassUser{}
									user.Parents = make([]dt.AddressBookClassParent, 0, 1)
									if infoNode != nil && infoNode.FirstChild != nil {
										userNode := infoNode.FirstChild
										for userNode != nil && userNode.Data != "td" && userNode.Data != "th" {
											userNode = userNode.NextSibling
										}
										if userNode != nil && userNode.FirstChild != nil {
											userNameNode := userNode.FirstChild
											for userNameNode != nil && userNameNode.Data != "a" {
												userNameNode = userNameNode.NextSibling
											}
											if userNameNode != nil && userNameNode.FirstChild != nil {
												for _, a := range userNameNode.Attr {
													if a.Key == "href" {
														var start, end, start2, end2 int
														j := 0
														for i := 0; i < len(a.Val); i++ {
															if a.Val[i:i+1] == "'" {
																if j == 0 {
																	start = i + 1
																	j++
																} else {
																	if j == 1 {
																		end = i
																		j++
																	} else {
																		if j == 2 {
																			start2 = i + 1
																			j++
																		} else {
																			end2 = i
																			break
																		}
																	}

																}
															}
														}
														user.StudentID = a.Val[start:end]
														user.Student = a.Val[start2:end2]
														break
													}
												}
											}
										}
										if userNode != nil && userNode.NextSibling != nil {
											userNode = userNode.NextSibling
											for userNode != nil && userNode.Data != "td" && userNode.Data != "th" {
												userNode = userNode.NextSibling
											}
											if userNode != nil && userNode.FirstChild != nil && userNode.FirstChild.FirstChild != nil {
												userNode = userNode.FirstChild
												for userNode != nil && userNode.Data != "a" {
													userNode = userNode.NextSibling
												}
												if userNode != nil {
													for userNode != nil {
														if userNode.FirstChild != nil {
															parent := dt.AddressBookClassParent{}
															for _, a := range userNode.Attr {
																if a.Key == "href" {
																	var start, end, start2, end2 int
																	j := 0
																	for i := 0; i < len(a.Val); i++ {
																		if a.Val[i:i+1] == "'" {
																			if j == 0 {
																				start = i + 1
																				j++
																			} else {
																				if j == 1 {
																					end = i
																					j++
																				} else {
																					if j == 2 {
																						start2 = i + 1
																						j++
																					} else {
																						end2 = i
																						break
																					}
																				}

																			}
																		}
																	}
																	parent.ParentID = a.Val[start:end]
																	parent.Parent = a.Val[start2:end2]
																	user.Parents = append(user.Parents, parent)
																	break
																}
															}
														}

														userNode = userNode.NextSibling
													}
												}
											}
										}
										class.Users = append(class.Users, user)
									}
								}
							}
						}
					}
				}
			}
			return nil, &class, nil
		}

		// В html-коде записан класс, составляем структуру AddressBookGroup
		group := dt.AddressBookGroup{}
		group.Title = groupName
		if infoNode.NextSibling != nil {
			infoNode = infoNode.NextSibling
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
					for infoNode != nil && infoNode.Data != "table" {
						infoNode = infoNode.NextSibling
					}
					if infoNode != nil && infoNode.FirstChild != nil {
						infoNode = infoNode.FirstChild
						for infoNode != nil && infoNode.Data != "tbody" {
							infoNode = infoNode.NextSibling
						}
						if infoNode != nil && infoNode.FirstChild != nil {
							infoNode = infoNode.FirstChild
							for infoNode != nil && infoNode.Data != "tr" {
								infoNode = infoNode.NextSibling
							}
							if infoNode != nil && infoNode.FirstChild != nil {
								infoNode = infoNode.FirstChild
								for infoNode != nil && infoNode.Data != "td" {
									infoNode = infoNode.NextSibling
								}
								if infoNode != nil && infoNode.FirstChild != nil {
									group.Users = make([]dt.AddressBookGroupUser, 0, 5)
									for infoNode = infoNode.FirstChild; infoNode != nil; infoNode = infoNode.NextSibling {
										if infoNode.FirstChild != nil {
											user := dt.AddressBookGroupUser{}
											user.Name = infoNode.FirstChild.Data
											for _, a := range infoNode.Attr {
												if a.Key == "href" {
													var start, end int
													j := 0
													for i := 0; i < len(a.Val); i++ {
														if a.Val[i:i+1] == "'" {
															if j == 0 {
																start = i + 1
																j++
															} else {
																end = i
																break
															}
														}
													}
													user.ID, err = strconv.Atoi(a.Val[start:end])
												}
											}
											group.Users = append(group.Users, user)
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return &group, nil, nil
	}

	// Возвращает группу или класс в адресной книге в зависимости от того, что есть в переданном html-коде
	parseAddressBookTable := func(node *html.Node) (*dt.AddressBookGroup, *dt.AddressBookClass, error) {
		tableNode := findAddressBookTableNode(node)
		return formAddressBookInfo(tableNode)
	}
	return parseAddressBookTable(parsedHTML)
}
