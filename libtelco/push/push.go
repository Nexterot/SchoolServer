/*
Package push содержит объявления функций, посылающих пуши.
*/
package push

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	api "github.com/masyagin1998/SchoolServer/libtelco/rest-api"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	db "github.com/masyagin1998/SchoolServer/libtelco/sql-db"
	"github.com/pkg/errors"
)

// Push struct содержит конфигурацию пушей.
type Push struct {
	api     *api.RestAPI
	db      *db.Database
	logger  *log.Logger
	stopped bool
	period  time.Duration
}

// NewPush создает структуру пушей и возвращает указатель на неё.
func NewPush(restapi *api.RestAPI, logger *log.Logger) *Push {
	return &Push{
		api:     restapi,
		db:      restapi.Db,
		logger:  logger,
		stopped: true,
		period:  time.Second * 30,
	}
}

// Run запускает функцию рассылки пушей.
func (p *Push) Run() {
	p.logger.Info("PUSH: Started")
	p.stopped = false
	// В бесконечном цикле с заданным периодом пускать горутину
	for {
		timer := time.NewTimer(p.period)
		<-timer.C
		if p.stopped == false {
			go p.sendPushes()
		}
	}
}

// Stop останавливает рассылку.
func (p *Push) Stop() {
	p.logger.Info("PUSH: Stoppped")
	p.stopped = true
}

// sendPushes рассылает пуши.
func (p *Push) sendPushes() {
	var (
		users   []db.User
		devices []db.Device
		school  db.School
	)
	// Пока обновляться будут только оценки
	p.logger.Info("PUSH: Sending push notifications")
	// shortcut
	pg := p.db.SchoolServerDB
	// Достанем всех пользователей
	err := pg.Find(&users).Error
	if err != nil {
		p.logger.Error("PUSH: Error when getting users list", "Error", err)
		return
	}
	// Гоним по пользователям
	for _, usr := range users {
		p.logger.Info("PUSH: user", "Login", usr.Login)
		// Получаем школу по id
		err := pg.First(&school, usr.SchoolID).Error
		if err != nil {
			p.logger.Error("PUSH: Error when getting school by id", "Error", err, "SchoolID", usr.SchoolID)
			return
		}
		// Сходим за оценками на удаленный сервер
		config := cp.School{Link: school.Address, Login: usr.Login, Password: usr.Password, Type: school.Type}
		session := ss.NewSession(&config)
		// Залогинимся
		err = session.Login()
		if err != nil {
			p.logger.Error("PUSH: Error when logging in", "Error", err)
			return
		}
		// Получим ChidlrenMap
		err = session.GetChildrenMap()
		if err != nil {
			p.logger.Error("PUSH: Error when getting children map", "Error", err)
			return
		}
		// Получим дату из текущей недели
		//today := "21.05.2018"
		today := time.Now().AddDate(0, 0, -95).Format("02.01.2006")
		p.logger.Info(today)
		// Получим дату из следующей недели
		nextweek := time.Now().AddDate(0, 0, -88).Format("02.01.2006")
		p.logger.Info(nextweek)
		// Счетчики оценок
		totalChangedMarks := 0
		totalNewMarks := 0
		totalImportantChangedMarks := 0
		totalImportantNewMarks := 0
		totalNewTasks := 0
		totalNewHomeTasks := 0
		// Гоним по ученикам пользователя
		for _, stud := range session.Children {
			// Вызовем GetWeekSchoolMarks для текущей недели
			marks, err := session.GetWeekSchoolMarks(today, stud.SID)
			if err != nil {
				p.logger.Error("PUSH: Error when getting marks", "Error", err, "Date", today, "StudentID", stud.SID)
				return
			}
			// Сравним с версией из БД
			chs, err := p.countChanges(usr.ID, stud.SID, marks)
			if err != nil {
				p.logger.Error("PUSH: Error when getting marks from db", "Error", err)
				return
			}
			totalChangedMarks += chs.nChangedMarks
			totalNewMarks += chs.nNewMarks
			totalImportantChangedMarks += chs.nImportantChangedMarks
			totalImportantNewMarks += chs.nImportantNewMarks
			totalNewTasks += chs.nNewTasks
			totalNewHomeTasks += chs.nNewHomeTasks
			// Вызовем GetWeekSchoolMarks для следующей недели
			marks, err = session.GetWeekSchoolMarks(nextweek, stud.SID)
			if err != nil {
				p.logger.Error("PUSH: Error when getting marks", "Error", err, "Date", nextweek, "StudentID", stud.SID)
				return
			}
			// Сравним с версией из БД
			chs, err = p.countChanges(usr.ID, stud.SID, marks)
			if err != nil {
				p.logger.Error("PUSH: Error when getting marks from db", "Error", err)
				return
			}
			totalChangedMarks += chs.nChangedMarks
			totalNewMarks += chs.nNewMarks
			totalImportantChangedMarks += chs.nImportantChangedMarks
			totalImportantNewMarks += chs.nImportantNewMarks
			totalNewTasks += chs.nNewTasks
			totalNewHomeTasks += chs.nNewHomeTasks
		}
		// Выйдем из системы
		if err := session.Logout(); err != nil {
			p.logger.Error("PUSH: Error when logging out", "Error", err)
			return
		}
		// debug
		p.logger.Info("PUSH: marks", "totalChanged", totalChangedMarks, "totalNew", totalNewMarks, "totalImportantChanged", totalImportantChangedMarks, "totalImportantNew", totalImportantNewMarks)
		p.logger.Info("PUSH: tasks", "totalNewTasks", totalNewTasks, "totalNewHomeTasks", totalNewHomeTasks)
		// Достанем все девайсы пользователя
		err = pg.Model(&usr).Related(&devices).Error
		if err != nil {
			p.logger.Error("PUSH: Error when getting devices list", "Error", err, "User", usr)
			return
		}
		// Гоним по девайсам
		for _, dev := range devices {
			p.logger.Info("PUSH: device", "System", dev.SystemType, "Token", dev.Token, "MarksNotification", dev.MarksNotification)
			// Если андроид
			if dev.SystemType == db.Android {
				// Посмотрим, каким образом надо выводить уведомления
				var n, k, d int
				if dev.TasksNotification == db.TasksNotificationAll {
					n = totalNewTasks
				} else if dev.TasksNotification == db.TasksNotificationHome {
					n = totalNewHomeTasks
				} else {
					n = 0
				}
				if dev.MarksNotification == db.MarksNotificationAll {
					k = totalNewMarks
					d = totalChangedMarks
				} else if dev.MarksNotification == db.MarksNotificationImportant {
					k = totalImportantNewMarks
					d = totalImportantChangedMarks
				} else {
					k = 0
					d = 0
				}
				if n+k+d > 3 {
					// У вас n новых заданий, k новых оценок, d измененных оценок
					msg := "У Вас "
					if n > 0 {
						msg += strconv.Itoa(n)
						// Посклоняем слова
						if (n % 10) == 0 {
							msg += " новых заданий"
						} else if (n % 10) == 1 {
							msg += " новое задание"
						} else if (n % 10) > 4 {
							msg += " новых заданий"
						} else {
							msg += " новых задания"
						}
					}
					if k > 0 {
						if n > 0 {
							msg += ", "
						}
						msg += strconv.Itoa(k)
						// Посклоняем слова
						if (k % 10) == 0 {
							msg += " новых оценок"
						} else if (k % 10) == 1 {
							msg += " новая оценка"
						} else if (k % 10) > 4 {
							msg += " новых оценок"
						} else {
							msg += " новых оценки"
						}
					}
					if d > 0 {
						if n+k > 0 {
							msg += ", "
						}
						msg += strconv.Itoa(d)
						// Посклоняем слова
						if (d % 10) == 0 {
							msg += " изменённых оценок"
						} else if (d % 10) == 1 {
							msg += " изменённая оценка"
						} else if (d % 10) > 4 {
							msg += " изменённых оценок"
						} else {
							msg += " изменённых оценки"
						}
					}
					// Отправить пуш
					err = p.sendPush(msg, dev.SystemType, dev.Token)
					if err != nil {
						p.logger.Error("PUSH: Error when sending push to client", "Error", err, "msg", msg, "Platform Type", dev.SystemType, "Token", dev.Token)
						return
					}
					// Появилось s новых объявлений
					// У вас w новых писем на почте
				} else {
					// У вас новая оценка (2 новые оценки)
					if k > 0 {
						msg := "У Вас "
						if k == 1 {
							msg += "новая оценка"
						} else {
							msg += strconv.Itoa(k)
							msg += " новые оценки"
						}
						// Отправить пуш
						err = p.sendPush(msg, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "msg", msg, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
					// У вас изменена оценка (изменены 2 оценки)
					if d > 0 {
						msg := "У Вас "
						if d == 1 {
							msg += "изменена оценка"
						} else {
							msg += " изменены "
							msg += strconv.Itoa(d)
							msg += " оценки"
						}
						// Отправить пуш
						err = p.sendPush(msg, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "msg", msg, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
					// У вас новое домашнее задание (2 новых домашних задания)
					if n > 0 {
						msg := "У Вас "
						if totalNewHomeTasks == 1 {
							msg += "новое домашнее задание"
						} else {
							msg += strconv.Itoa(totalNewHomeTasks)
							msg += " новых домашних задания"
						}
						// Отправить пуш
						err = p.sendPush(msg, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "msg", msg, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
					// У вас новая работа (2 новые работы)
					if n > 0 && dev.TasksNotification == db.TasksNotificationAll {
						msg := "У Вас "
						n -= totalNewHomeTasks
						if n == 1 {
							msg += "новая работа"
						} else {
							msg += strconv.Itoa(n)
							msg += " новые работы"
						}
						// Отправить пуш
						err = p.sendPush(msg, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "msg", msg, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
					// Появилось новое объявление (2 новых объявления)
					// У вас новое сообщение на почте (2 новых сообщения)
				}
			}
		}
	}
}

// changes struct содержит количество изменений для отправления с помощью push.
type changes struct {
	// Оценки
	nChangedMarks          int
	nNewMarks              int
	nImportantChangedMarks int
	nImportantNewMarks     int
	// Задания
	nNewTasks     int
	nNewHomeTasks int
}

// countChanges считает количество изменений.
func (p *Push) countChanges(userID uint, studentID string, week *dt.WeekSchoolMarks) (*changes, error) {
	var (
		student db.Student
		days    []db.Day
		tasks   []db.Task
		newDay  db.Day
		newTask db.Task
		chs     changes
	)
	id, err := strconv.Atoi(studentID)
	if err != nil {
		return nil, errors.Wrap(err, "PUSH: Error when converting studentID")
	}
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем ученика по userID и studentID
	where := db.Student{NetSchoolID: id, UserID: userID}
	err = pg.Where(where).First(&student).Error
	if err != nil {
		return nil, errors.Wrap(err, "PUSH: Error when getting student")
	}
	// Получаем список дней у ученика
	err = pg.Model(&student).Related(&days).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting student='%v' days", student)
	}
	// Гоняем по дням из пакета
	for dayNum, day := range week.Data {
		date := day.Date
		// Найдем подходящий день в БД
		dbDayFound := false
		for _, dbDay := range days {
			if date == dbDay.Date {
				dbDayFound = true
				newDay = dbDay
				break
			}
		}
		if !dbDayFound {
			// Дня не существует, надо создать
			newDay = db.Day{StudentID: student.ID, Date: date, Tasks: []db.Task{}}
			err = pg.Create(&newDay).Error
			if err != nil {
				return nil, errors.Wrapf(err, "Error creating newDay='%v'", newDay)
			}
			days = append(days, newDay)
		}
		// Получаем список заданий для дня
		err = pg.Model(&newDay).Related(&tasks).Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting newDay='%v' tasks", newDay)
		}
		// Гоняем по заданиям
		for taskNum, task := range day.Lessons {
			// Найдем подходящее задание в БД
			dbTaskFound := false
			for _, dbTask := range tasks {
				if task.AID == dbTask.AID {
					dbTaskFound = true
					newTask = dbTask
					// Сравнить оценки
					if task.Mark != dbTask.Mark {
						// Если оценки не совпали
						if dbTask.Mark == "-" {
							// Если в БД лежит пустая оценка, значит оценка новая
							if task.Type == "В" || task.Type == "К" {
								// Если срезовая оценка или контрольная, она важна
								chs.nImportantNewMarks++
							}
							chs.nNewMarks++
						} else {
							if task.Type == "В" || task.Type == "К" {
								// Если срезовая оценка или контрольная, она важна
								chs.nImportantChangedMarks++
							}
							// Иначе оценка была изменена
							chs.nChangedMarks++
						}
						dbTask.Mark = task.Mark
						err = pg.Save(&dbTask).Error
						if err != nil {
							return nil, errors.Wrapf(err, "Error saving newTask='%v'", newTask)
						}
					}
					break
				}
			}
			if !dbTaskFound {
				// Задания не существует, надо создать
				newTask = db.Task{DayID: newDay.ID, CID: task.CID, AID: task.AID, Status: db.StatusTaskNew, TP: task.TP, InTime: task.InTime, Name: task.Name,
					Title: task.Title, Type: task.Type, Mark: task.Mark, Weight: task.Weight, Author: task.Author}
				err = pg.Create(&newTask).Error
				if err != nil {
					return nil, errors.Wrapf(err, "Error creating newTask='%v'", newTask)
				}
				tasks = append(tasks, newTask)
				// Новое задание, запишем в счетчик
				if newTask.Type == "Д" {
					// Если домашняя работа, так же обновим счетчик
					chs.nNewHomeTasks++
				}
				chs.nNewTasks++
			}
			// Присвоить статусу таска из пакета статус таска из БД
			week.Data[dayNum].Lessons[taskNum].Status = newTask.Status
		}
		err = pg.Save(&newDay).Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error saving newDay='%v'", newDay)
		}
	}
	err = pg.Save(&student).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error saving student='%v'", student)
	}
	return &chs, nil
}

type gorushRequest struct {
	Notifications []notification `json:"notifications"`
}

type notification struct {
	Tokens   []string `json:"tokens"`
	Platform int      `json:"platform"`
	Message  string   `json:"message"`
}

// sendPush посылает post-запрос на webapi gorush.
func (p *Push) sendPush(msg string, platformType int, token string) error {
	var (
		tokens        []string
		notifications []notification
	)
	tokens = append(tokens, token)
	notifications = append(notifications, notification{Tokens: tokens, Platform: platformType, Message: msg})
	req := gorushRequest{Notifications: notifications}
	url := "http://localhost:8088/api/push"
	byt, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "PUSH: Error marshalling")
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(byt))
	if err != nil {
		return errors.Wrap(err, "PUSH: Error sending web api gorush request")
	}
	defer resp.Body.Close()
	p.logger.Info("PUSH: Got response from gorush", "Response", resp)
	return nil
}
