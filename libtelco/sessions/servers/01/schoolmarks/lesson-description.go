// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package schoolmarks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	gr "github.com/levigross/grequests"
	redis "github.com/masyagin1998/SchoolServer/libtelco/in-memory-db"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetLessonDescription вовзращает подробности урока с сервера первого типа.
func GetLessonDescription(s *dt.Session, AID, CID, TP int, schoolID, studentID, classID, serverAddr string, db *redis.Database) (*dt.LessonDescription, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AID":       strconv.Itoa(AID),
				"AT":        s.AT,
				"CID":       strconv.Itoa(CID),
				"PCLID_IUP": "",
				"TP":        strconv.Itoa(TP),
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Curriculum/Assignments.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/Assignments/GetAssignmentInfo.asp", ro)
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
	// находящуюся в теле ответа, и найти в ней подробности урока.

	// Получаем таблицу с подробностями урока из полученной json-структуры
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(b, &responseMap)
	if err != nil {
		return nil, err
	}
	strs := responseMap["data"].(map[string]interface{})
	var responseString string
	var authorString string
	for k, v := range strs {
		if k == "strTable" {
			responseString = v.(string)
		}
		if k == "strTitle" {
			authorString = v.(string)
			var start, end int
			for i := 0; i < len(authorString); i++ {
				if authorString[i:i+1] == "(" {
					start = i + 1
				} else {
					if authorString[i:i+1] == ")" {
						end = i
					}
				}
			}
			authorString = authorString[start:end]
		}
	}

	parsedHTML, err := html.Parse(strings.NewReader(responseString))
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findLessonDescriptionTableNode func(*html.Node) *html.Node
	findLessonDescriptionTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "table table-bordered table-condensed" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findLessonDescriptionTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит из переданной строки путь к файлу и его ID
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

	formLessonDescription := func(node *html.Node) (*dt.LessonDescription, string, error) {
		var ID string
		details := *new(dt.LessonDescription)
		if node != nil {

			var authorNode *html.Node
			if node.Parent != nil && node.Parent.Parent != nil && node.Parent.Parent.Parent != nil && node.Parent.Parent.Parent.PrevSibling != nil {
				authorNode = node.Parent.Parent.Parent.PrevSibling
				if authorNode.FirstChild != nil && authorNode.FirstChild.FirstChild != nil && authorNode.FirstChild.FirstChild.NextSibling != nil {
					authorNode = authorNode.FirstChild.FirstChild.NextSibling
					if authorNode.FirstChild != nil {
						author := authorNode.FirstChild.Data
						var start, end int
						for i := 0; i < len(author); i++ {
							if author[i:i+1] == "(" {
								start = i + 1
							} else {
								if author[i:i+1] == ")" {
									end = i
								}
							}
						}
						details.Author = author[start:end]
					}
				}
			}

			tableNode := node.FirstChild.FirstChild.NextSibling.NextSibling
			commentNode := tableNode.FirstChild.NextSibling
			if commentNode.FirstChild != nil {
				commentNode = commentNode.FirstChild
				for commentNode != nil {
					if commentNode.FirstChild != nil {
						details.Description += commentNode.FirstChild.Data
					} else {
						if commentNode.Data == "br" {
							details.Description += "\n"
						} else {
							if !(len(commentNode.Data) == 2 && commentNode.Data[0] == 194 && commentNode.Data[1] == 160) {
								details.Description += commentNode.Data
							}
						}
					}

					commentNode = commentNode.NextSibling
				}
			}

			tableNode = tableNode.NextSibling
			if tableNode.FirstChild.NextSibling.FirstChild != nil {
				if tableNode.FirstChild.NextSibling.FirstChild.FirstChild != nil {
					if tableNode.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild != nil {
						details.FileName = tableNode.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild.Data
					}
					for _, a := range tableNode.FirstChild.NextSibling.FirstChild.FirstChild.Attr {
						if a.Key == "href" {
							details.File, ID, err = findURLAndID(a.Val)
							if err != nil {
								return &details, ID, err
							}
							break
						}
					}
				}
			}

			return &details, ID, nil
		}

		return &details, ID, errors.New("Node is nil in func formLessonDescription")
	}

	makeLessonDescription := func(node *html.Node) (*dt.LessonDescription, string, error) {
		var details *dt.LessonDescription
		var ID string
		tableNode := findLessonDescriptionTableNode(node)
		details, ID, err = formLessonDescription(tableNode)

		return details, ID, err
	}

	lessonDesc, ID, err := makeLessonDescription(parsedHTML)
	if lessonDesc != nil && authorString != "" {
		lessonDesc.Author = authorString
	}
	if err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно запустить закачку файла и
	// подменить нашей ссылкой ссылку NetSchool.
	return getFile(s, lessonDesc, schoolID, classID, ID, serverAddr, db)
}

// getFile выкачивает файл по заданной ссылке в заданную директорию (если его там ещё нет) и возвращает
// - true, если файл был скачан;
// - false, если файл уже был в директории;
// с сервера первого типа.
func getFile(s *dt.Session, lessonDesc *dt.LessonDescription, schoolID, classID, ID, serverAddr string, db *redis.Database) (*dt.LessonDescription, error) {
	p := "http://"

	// Проверка, а есть ли вообще прикрепленный файл.
	if lessonDesc.FileName == "" {
		return lessonDesc, nil
	}

	// Проверка, есть ли файл на диске.
	path := fmt.Sprintf("files/%s/classes/%s/%s/", schoolID, classID, ID)
	if _, err := os.Stat(path + lessonDesc.FileName); err == nil {
		// Проверка, актуален ли он (время "протухания" файла - 12 часов).
		stringTime, err := db.GetFileDate(path + lessonDesc.FileName)
		if err != nil {
			return nil, err
		}
		fileTime, err := strconv.ParseInt(stringTime, 10, 64)
		if err != nil {
			return nil, err
		}
		currTime := time.Now().Unix()
		if (currTime - fileTime) < 12*3600 {
			lessonDesc.File = serverAddr + "/doc/" + path + lessonDesc.FileName
			return lessonDesc, nil
		}
	}

	// Закачка файла.

	// 0-ой POST-запрос.
	ro := &gr.RequestOptions{
		Data: map[string]string{
			"VER":          s.VER,
			"at":           s.AT,
			"attachmentId": ID,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Curriculum/Assignments.asp",
		},
	}
	r, err := s.Sess.Post(p+s.Serv.Link+lessonDesc.File, ro)
	if err != nil {
		lessonDesc.File = ""
		lessonDesc.FileName = "Broken"
		return lessonDesc, nil
	}
	defer func() {
		_ = r.Close()
	}()
	if _, err := check.CheckResponse(s, r); err != nil {
		lessonDesc.File = ""
		lessonDesc.FileName = "Broken"
		return lessonDesc, nil
	}
	// Сохранение файла на диск.
	// Создание папок, если надо.
	if err = os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	// Создание файла.
	if err := ioutil.WriteFile(path+lessonDesc.FileName, r.Bytes(), 0700); err != nil {
		return nil, err
	}
	// Подмена Ссылки.
	lessonDesc.File = serverAddr + "/doc/" + path[6:] + lessonDesc.FileName
	// Запись в Redis.
	if err := db.AddFileDate(path+lessonDesc.FileName, fmt.Sprintf("%v", time.Now().Unix())); err != nil {
		return nil, err
	}
	return lessonDesc, nil
}
