package base

// GetChildrenMap получает мапу детей в их UID с сервера первого типа.
import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func GetChildrenMap(s *dt.Session) error {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "0",
				"ThmID":     "1",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", ro)
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
	// находящуюся в теле ответа, и найти в ней мапу детей в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "parsing HTML")
	}

	var getChildrenIDNode func(*html.Node) *html.Node
	getChildrenIDNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "select" || node.Data == "input" {
				for _, a := range node.Attr {
					if a.Key == "name" && a.Val == "SID" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := getChildrenIDNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит ID учеников и также определяет тип сессии
	getChildrenIDs := func(node *html.Node) (map[string]dt.Student, bool, error) {
		// Находим ID учеников/ученика.
		childrenIDs := make(map[string]dt.Student)
		idNode := getChildrenIDNode(node)
		if idNode != nil {
			if idNode.FirstChild == nil {
				for _, a := range idNode.Attr {
					if a.Key == "value" {
						if idNode.PrevSibling != nil {
							nameNode := idNode.PrevSibling
							flag := true
							for nameNode != nil && flag {
								for _, b := range nameNode.Attr {
									if b.Key == "type" && b.Val == "text" {
										flag = false
										break
									}
								}
								if !flag {
									break
								}
								nameNode = nameNode.PrevSibling
							}
							if nameNode != nil && !flag {
								for _, a2 := range nameNode.Attr {
									if a2.Key == "value" {
										childrenIDs[a2.Val] = dt.Student{a.Val, ""}
										if _, err := strconv.Atoi(a.Val); err != nil {
											return nil, false, fmt.Errorf("ID has incorrect format \"%v\"", a.Val)
										}
									}
								}
							}
						}
					}
				}
			} else {
				for n := idNode.FirstChild; n != nil; n = n.NextSibling {
					if len(n.Attr) != 0 {
						for _, a := range n.Attr {
							if a.Key == "value" {
								childrenIDs[n.FirstChild.Data] = dt.Student{a.Val, ""}
								if _, err := strconv.Atoi(a.Val); err != nil {
									return nil, false, fmt.Errorf("ID has incorrect format \"%v\"", a.Val)
								}
							}
						}
					}
				}
			}
		} else {
			return nil, false, fmt.Errorf("Couldn't find children IDs Node")
		}

		// Находим тип сессии.
		sessTypeNode := idNode.Parent
		if sessTypeNode != nil && sessTypeNode.Data == "select" {
			sessTypeNode = sessTypeNode.Parent
		}
		isParent := false
		for sessTypeNode != nil && sessTypeNode.Data != "label" {
			sessTypeNode = sessTypeNode.PrevSibling
		}
		if sessTypeNode != nil && sessTypeNode.FirstChild != nil {
			if sessTypeNode.FirstChild.Data == "Ученики" {
				isParent = true
			}
		} else {
			return nil, false, errors.New("Couldn't find type of session")
		}

		return childrenIDs, isParent, nil
	}

	var isParent bool
	s.Children, isParent, err = getChildrenIDs(parsedHTML)
	if err != nil {
		return errors.Wrap(err, "parsing")
	}
	if isParent {
		s.Type = dt.Parent
	} else {
		s.Type = dt.Child
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "3",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return check.CheckResponse(s, r)
	}

	type Filter struct {
		ID    string `json:"filterId"`
		Value string `json:"filterValue"`
	}

	type SelectedData struct {
		SelectedData []Filter `json:"selectedData"`
	}

	// 2-ой Post-запрос.
	r2 := func(json SelectedData) ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			JSON: json,
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/webapi/reports/journal_access/initfilters", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}

	// Если мы дошли до этого места, то можно начать искать CLID каждого ребенка.
	for k, v := range s.Children {
		flag, err := r1()
		if err != nil {
			return errors.Wrap(err, "1 POST")
		}
		if !flag {
			flag, err = r1()
			if err != nil {
				return errors.Wrap(err, "retrying 1 POST")
			}
			if !flag {
				return fmt.Errorf("retry didn't work for 1 POST")
			}
		}
		json := SelectedData{
			SelectedData: []Filter{Filter{"SID", v.SID}},
		}
		b, flag, err := r2(json)
		if err != nil {
			return err
		}
		if !flag {
			b, flag, err = r2(json)
			if err != nil {
				return err
			}
			if !flag {
				return fmt.Errorf("Retry didn't work")
			}
		}
		CLID := string(b)
		index := strings.Index(CLID, "\"value\":\"")
		if index == -1 {
			return fmt.Errorf("Invalid begin SID substring \"%s\"", CLID)
		}
		CLID = CLID[index+len("\"value\":\""):]
		index = strings.Index(CLID, "\"")
		if index == -1 {
			return fmt.Errorf("Invalid end SID substring \"%s\"", CLID)
		}
		CLID = CLID[:index]
		v.CLID = CLID
		s.Children[k] = v
	}

	if len(s.Children) == 1 {
		for _, v := range s.Children {
			s.Child = v
		}
	}
	return nil
}
