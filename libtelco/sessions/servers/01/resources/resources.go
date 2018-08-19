// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package resources

import (
	"bytes"
	"fmt"
	"strings"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// GetResourcesList возвращает список всех ресурсов с сервера первого типа.
func GetResourcesList(s *dt.Session) (*dt.Resources, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"MenuItem":  "0",
				"TabItem":   "40",
				"VER":       s.VER,
				"optional":  "optional",
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Curriculum/SchoolResources.asp", ro)
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
	// находящуюся в теле ответа, и найти в ней список ресурсов.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Находит все ноды, содержащие группу
	var getGroupsNodes func(*html.Node, *[]*html.Node)
	getGroupsNodes = func(node *html.Node, groupNodes *[]*html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "li" {
				for _, a := range node.Attr {
					if (a.Key == "class") && (a.Val == "tree-group") {
						*groupNodes = append(*groupNodes, node)
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			getGroupsNodes(c, groupNodes)
		}
	}

	// Находит все файлы из переданного нода
	makeFiles := func(node *html.Node) []dt.File {
		files := make([]dt.File, 0, 1)
		hasHref := true
		for fileNode := node.FirstChild; fileNode != nil; fileNode = fileNode.NextSibling {
			if hasHref {
				file := *new(dt.File)
				for _, a := range fileNode.Attr {
					if a.Key == "href" {
						file.Link = a.Val
						file.Name = fileNode.NextSibling.FirstChild.Data
						files = append(files, file)
						break
					}
				}
			}
			hasHref = !hasHref
		}
		return files
	}

	// Убирает все пробелы из строки
	deleteSpaces := func(str string) string {
		returnString := ""
		for i := 0; i < len(str); i++ {
			if str[i] != 9 && str[i] != 10 {
				returnString += str[i : i+1]
			}
		}
		return returnString
	}

	// Формирует все группы, которые находятся в переданных нодах
	formGroups := func(nodes []*html.Node) ([]dt.Group, error) {
		groups := make([]dt.Group, 0, len(nodes))
		if len(nodes) == 0 {
			return groups, errors.New("Couldn't find groups in html-code. Perhaps, there are no school resources on this page")
		}
		for i := 0; i < len(nodes); i++ {
			dataNode := nodes[i].FirstChild.NextSibling
			for dataNode != nil && dataNode.Data != "span" {
				dataNode = dataNode.NextSibling
			}
			group := *new(dt.Group)
			if dataNode.FirstChild.NextSibling.FirstChild != nil {
				group.GroupTitle = deleteSpaces(dataNode.FirstChild.NextSibling.FirstChild.Data)
			}
			for dataNode != nil && !((len(dataNode.Attr)) == 1 && dataNode.Attr[0].Val == "tree-group-item") {
				dataNode = dataNode.NextSibling
			}
			if dataNode != nil {
				dataNode = dataNode.FirstChild
				hasSubgroups := false
				for dataNode != nil {
					if len(dataNode.Attr) == 1 && dataNode.Attr[0].Val == "tree-group-item" {
						break
					}
					if len(dataNode.Attr) == 1 && strings.Contains(dataNode.Attr[0].Val, "tree-group-subgroup") {
						hasSubgroups = true
						break
					}
					dataNode = dataNode.NextSibling
				}
				if dataNode != nil {
					var subgroups []dt.Subgroup
					if hasSubgroups {
						subgroups = make([]dt.Subgroup, 0, 1)
						for subgroupNode := dataNode; subgroupNode != nil; subgroupNode = subgroupNode.NextSibling {
							if !(len(subgroupNode.Attr) == 1 && strings.Contains(subgroupNode.Attr[0].Val, "tree-group-subgroup")) {
								continue
							}
							c := subgroupNode.FirstChild
							for c != nil && c.Data != "span" {
								c = c.NextSibling
							}

							subgroup := *new(dt.Subgroup)
							if c.FirstChild.NextSibling.FirstChild != nil {
								subgroup.SubgroupTitle = deleteSpaces(c.FirstChild.NextSibling.FirstChild.Data)
							}
							c = c.NextSibling.NextSibling.FirstChild
							for c != nil && !(len(c.Attr) == 1 && c.Attr[0].Val == "tree-group-item") {
								c = c.NextSibling
							}
							if c != nil {
								c = c.FirstChild
								for c != nil && !(len(c.Attr) == 1 && c.Attr[0].Val == "tree-group-item") {
									c = c.NextSibling
								}
								if c != nil {
									c = c.FirstChild.NextSibling.NextSibling.NextSibling
									subgroup.Files = makeFiles(c)
								}
							}
							subgroups = append(subgroups, subgroup)
						}
						group.Files = make([]dt.File, 0)
					} else {
						subgroups = make([]dt.Subgroup, 0)
						dataNode = dataNode.FirstChild.NextSibling
						group.Files = makeFiles(dataNode.NextSibling.NextSibling)
					}
					group.Subgroups = subgroups
					groups = append(groups, group)
				}
			}
		}

		return groups, nil
	}

	// Создаёт отчёт
	makeSchoolResources := func(node *html.Node) (*dt.Resources, error) {
		var resources dt.Resources
		groupsNodes := make([]*html.Node, 0, 1)
		getGroupsNodes(node, &groupsNodes)
		resources.Data, err = formGroups(groupsNodes)

		return &resources, err
	}
	return makeSchoolResources(parsedHTML)
}
