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
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	db "github.com/masyagin1998/SchoolServer/libtelco/sql-db"
	"github.com/pkg/errors"
)

// Push struct содержит конфигурацию пушей.
type Push struct {
	db            *db.Database
	logger        *log.Logger
	stopped       bool
	Period        time.Duration
	GorushAddress string
	AppTopic      string
}

// NewPush создает структуру пушей и возвращает указатель на неё.
func NewPush(database *db.Database, logger *log.Logger) *Push {
	return &Push{
		db:            database,
		logger:        logger,
		stopped:       true,
		Period:        time.Second * 20,
		GorushAddress: "http://localhost:8088/api/push",
		AppTopic:      "kir4567.NetSchoolApp",
	}
}

// Run запускает функцию рассылки пушей.
func (p *Push) Run() {
	p.logger.Info("PUSH: Started")
	p.stopped = false
	// В бесконечном цикле с заданным периодом пускать горутину
	for {
		timer := time.NewTimer(p.Period)
		<-timer.C
		if p.stopped == false {
			go p.handlePushes()
		}
	}
}

// Stop останавливает рассылку.
func (p *Push) Stop() {
	p.logger.Info("PUSH: Stoppped")
	p.stopped = true
}

// handlePushes содержит основную логику пушей
func (p *Push) handlePushes() {
	var (
		users    []db.User
		devices  []db.Device
		school   db.School
		students []db.Student
	)
	p.logger.Info("PUSH: Sending push notifications")
	// shortcut
	pg := p.db.SchoolServerDB
	// Достанем пользователей
	err := pg.Find(&users).Error
	if err != nil {
		p.logger.Error("PUSH: Error when getting users list", "Error", err)
		return
	}
	// Отображение schoolID в число новых для этой школы
	nResources := make(map[uint]*resourcesChanges)
	// Текущее время
	now := time.Now()
	nowAsString := now.Format("02.01.2006")
	nextAsString := now.AddDate(0, 0, 7).Format("02.01.2006")
	// Гоним по пользователям
	for _, usr := range users {
		p.logger.Info("PUSH: user", "Login", usr.Login)
		// очистим поля
		students = []db.Student{}
		// Получаем школу по id
		err := pg.First(&school, usr.SchoolID).Error
		if err != nil {
			p.logger.Error("PUSH: Error when getting school by id", "Error", err, "SchoolID", usr.SchoolID)
			return
		}
		// Сходим на удаленный сервер
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

		// ИЗМЕНЕНИЯ
		scheduleChanged := false
		newTasksMarks := diaryNewTasksMarks{}

		if usr.Role == "Ученик" {
			// Достанем всех "учеников" пользователя
			err = pg.Model(&usr).Related(&students).Error
			if err != nil {
				p.logger.Error("PUSH: Error when getting students", "Error", err, "User", usr)
				return
			}

			// Скачаем расписание
			week, err := session.GetTimeTable(nowAsString, 7, strconv.Itoa(students[0].NetSchoolID))
			if err != nil {
				p.logger.Error("PUSH: Error when getting schedule", "Error", err)
				return
			}
			scheduleChanged, err = p.checkSchedule(students[0].ID, week)
			if err != nil {
				p.logger.Error("PUSH: Error when checking schedule", "Error", err)
				return
			}

			// Дневник на текущую и следующие недели
			// текущая
			thisWeek, err := session.GetWeekSchoolMarks(nowAsString, strconv.Itoa(students[0].NetSchoolID))
			if err != nil {
				p.logger.Error("PUSH: Error when getting tasks and marks", "Error", err)
				return
			}
			err = p.checkDiary(students[0].ID, thisWeek, &newTasksMarks)
			if err != nil {
				p.logger.Error("PUSH: Error when checking tasks and marks", "Error", err)
				return
			}
			// следующая
			nextWeek, err := session.GetWeekSchoolMarks(nextAsString, strconv.Itoa(students[0].NetSchoolID))
			if err != nil {
				p.logger.Error("PUSH: Error when getting tasks and marks", "Error", err)
				return
			}
			err = p.checkDiary(students[0].ID, nextWeek, &newTasksMarks)
			if err != nil {
				p.logger.Error("PUSH: Error when checking tasks and marks", "Error", err)
				return
			}
		}

		// Скачаем форум
		forumTopics, err := session.GetForumThemesList("1")
		if err != nil {
			p.logger.Error("PUSH: Error when getting forum", "Error", err)
			return
		}
		err = p.checkForumTopics(usr.ID, forumTopics)
		if err != nil {
			p.logger.Error("PUSH: Error when checking forum topics", "Error", err)
			return
		}
		nForum := forumNewMessages{}
		for _, post := range forumTopics.Posts {
			// Для каждой темы скачаем сообщения
			messages, err := session.GetForumThemeMessages(strconv.Itoa(post.ID), "1", "10")
			if err != nil {
				p.logger.Error("PUSH: Error when getting forum messages", "Error", err)
				return
			}
			err = p.checkForumMessages(usr.ID, post.ID, post.Title, messages, &nForum)
			if err != nil {
				p.logger.Error("PUSH: Error when checking forum messages", "Error", err)
				return
			}
		}

		// Скачаем почту
		mailMessages, err := session.GetEmailsList("1", "0", "25", "DESC")
		if err != nil {
			p.logger.Error("PUSH: Error when getting mail", "Error", err)
			return
		}
		newMail := mailNewMessages{}
		err = p.checkMail(usr.ID, mailMessages, &newMail)
		if err != nil {
			p.logger.Error("PUSH: Error when checking mail", "Error", err)
			return
		}

		// С помощью первого пользователя из каждой школы скачаем ресурсы
		_, ok := nResources[usr.SchoolID]
		if !ok {
			resources, err := session.GetResourcesList()
			if err != nil {
				p.logger.Error("PUSH: Error when getting resource list", "Error", err)
				return
			}
			rChanges, err := p.checkResources(usr.SchoolID, resources)
			p.logger.Info("kek", "chan", rChanges)
			if err != nil {
				p.logger.Error("PUSH: Error when checking for new resources", "Error", err)
				return
			}
			nResources[usr.SchoolID] = rChanges
		}

		// Скачаем объявления
		posts, err := session.GetAnnouncements()
		if err != nil {
			p.logger.Error("PUSH: Error when getting posts", "Error", err)
			return
		}
		newPs := newPosts{}
		err = p.checkPosts(usr.ID, posts, &newPs)
		if err != nil {
			p.logger.Error("PUSH: Error when checking posts", "Error", err)
			return
		}

		// Выйдем из системы
		if err := session.Logout(); err != nil {
			p.logger.Error("PUSH: Error when logging out", "Error", err)
			return
		}

		// Достанем все девайсы пользователя
		err = pg.Model(&usr).Related(&devices).Error
		if err != nil {
			p.logger.Error("PUSH: Error when getting devices list", "Error", err, "User", usr)
			return
		}

		// Гоним по девайсам
		for _, dev := range devices {
			p.logger.Info("PUSH: device", "System", dev.SystemType, "Token", dev.Token)
			// Если стоит "не беспокоить", пропустим
			if dev.DoNotDisturbUntil != nil && now.Sub(*dev.DoNotDisturbUntil).String()[0] == '-' {
				p.logger.Info("Not disturbing this device until date", "Date", dev.DoNotDisturbUntil)
				continue
			}
			// Появился новый учебный материал
			rChanges := nResources[usr.SchoolID]
			p.logger.Info("Resources", "Number of changes", rChanges)
			if rChanges != nil && dev.ReportsNotification {
				for _, v := range rChanges.Changes {
					title := v.Title
					subtitle := v.Subtitle
					body := v.Body
					if v.Subtitle == "" {
						// это группа
						err = p.send(dev.SystemType, dev.Token, "resources_new_file_group", title, subtitle, body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					} else {
						// это файл
						err = p.send(dev.SystemType, dev.Token, "resources_new_file", title, subtitle, body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
				}
			}
			// Изменения в расписании
			p.logger.Info("Schedule", "Was Changed", scheduleChanged)
			if scheduleChanged && dev.ScheduleNotification {
				err = p.send(dev.SystemType, dev.Token, "schedule_change", "Изменения в расписании (см. детали)", "", "", "")
				if err != nil {
					p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
					return
				}
			}
			// Новые сообщения на форуме
			p.logger.Info("Forum", "Number of new messages", len(nForum.Messages))
			if dev.ForumNotification {
				if len(nForum.Messages) > 3 {
					err = p.send(dev.SystemType, dev.Token, "forum_new_message", "Форум", "", "Оставлено "+strconv.Itoa(len(nForum.Messages))+" новых сообщений", "")
					if err != nil {
						p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
						return
					}
				} else {
					for _, post := range nForum.Messages {
						err = p.send(dev.SystemType, dev.Token, "forum_new_message", post.Title, post.Subtitle, post.Body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
				}
			}
			// Новое почтовое сообщение
			p.logger.Info("Mail", "Number of new messages", len(newMail.Messages))
			if dev.MailNotification {
				if len(newMail.Messages) > 3 {
					err = p.send(dev.SystemType, dev.Token, "mail_new_message", "Почта", "", "У вас "+strconv.Itoa(len(newMail.Messages))+" новых сообщений", "")
					if err != nil {
						p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
						return
					}
				} else {
					for _, post := range newMail.Messages {
						err = p.send(dev.SystemType, dev.Token, "mail_new_message", post.Title, "", post.Body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
				}
			}
			// Новое задание или оценка
			p.logger.Info("Tasks and marks", "Total changes", len(newTasksMarks.TasksMarks))
			for _, task := range newTasksMarks.TasksMarks {
				if task.Type == Mark {
					switch dev.MarksNotification {
					case db.MarksNotificationDisabled:
						continue
					case db.MarksNotificationImportant:
						if task.IsImportant {
							err = p.send(dev.SystemType, dev.Token, "diary_new_mark", task.Title, "", task.Body, "")
							if err != nil {
								p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
								return
							}
						}
					case db.MarksNotificationAll:
						err = p.send(dev.SystemType, dev.Token, "diary_new_mark", task.Title, "", task.Body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					default:
						p.logger.Error("PUSH: Invalid mark type", "Type", task.Type)
						return
					}
				} else if task.Type == Task {
					switch dev.TasksNotification {
					case db.TasksNotificationDisabled:
						continue
					case db.TasksNotificationHome:
						if task.IsHomework {
							err = p.send(dev.SystemType, dev.Token, "diary_new_task", task.Title, "", task.Body, "")
							if err != nil {
								p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
								return
							}
						}
					case db.TasksNotificationAll:
						err = p.send(dev.SystemType, dev.Token, "diary_new_task", task.Title, "", task.Body, "")
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					default:
						p.logger.Error("PUSH: Invalid task type", "Type", task.Type)
						return
					}
				} else {
					p.logger.Error("PUSH: Invalid task or mark type", "Type", task.Type)
					return
				}
			}
			// Новое объявление
			p.logger.Info("Posts", "Total changes", len(newPs.Posts))
			if dev.ReportsNotification {
				for _, post := range newPs.Posts {
					err = p.send(dev.SystemType, dev.Token, "new_post", post.Title, "", post.Body, "")
					if err != nil {
						p.logger.Error("PUSH: Error when sending push to client", "Error", err, "Platform Type", dev.SystemType, "Token", dev.Token)
						return
					}

				}
			}
		}
	}
}

// newPosts struct
type newPosts struct {
	Posts []newPostStruct
}

// newPostStruct struct
type newPostStruct struct {
	Title string
	Body  string
}

// checkPosts
func (p *Push) checkPosts(userID uint, ps *dt.Posts, res *newPosts) error {
	var (
		user    db.User
		newPost db.Post
		posts   []db.Post
	)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем пользователя по pk userID
	err := pg.First(&user, userID).Error
	if err != nil {
		return errors.Wrap(err, "PUSH: Error when getting user")
	}
	// Получаем список объявлений у пользователя
	err = pg.Model(&user).Related(&posts).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' posts", user)
	}
	// Гоняем по объявлениям из пакета
	for _, post := range ps.Posts {
		// Найдем подходящую тему в БД
		postFound := false
		for _, dbPost := range posts {
			if post.Author == dbPost.Author && post.Title == dbPost.Title && post.Date == dbPost.Date {
				postFound = true
				newPost = dbPost
				break
			}
		}
		if !postFound {
			// Объявления не существует, надо создать
			newPost = db.Post{UserID: user.ID, Unread: post.Unread, Author: post.Author, Title: post.Title, Date: post.Date, Message: post.Message, File: post.FileLink, FileName: post.FileName}
			err = pg.Create(&newPost).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newPost='%v'", newPost)
			}
			posts = append(posts, newPost)
			// Запишем объявление в структуру
			res.Posts = append(res.Posts, newPostStruct{Title: newPost.Title, Body: newPost.Message})
		}
	}
	// Сохраним пользователя
	err = pg.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}

// diaryNewTasksMarks
type diaryNewTasksMarks struct {
	TasksMarks []diaryNewTaskMark
}

// Тип: задание или оценка
const (
	_ = iota
	Task
	Mark
)

// diaryNewTaskMark
type diaryNewTaskMark struct {
	Title       string
	Body        string
	IsImportant bool
	IsHomework  bool
	Type        int
}

// checkDiary
func (p *Push) checkDiary(studentID uint, week *dt.WeekSchoolMarks, tasksMarks *diaryNewTasksMarks) error {
	var (
		student db.Student
		days    []db.Day
		tasks   []db.Task
		newDay  db.Day
		newTask db.Task
	)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем ученика по pk studentID
	err := pg.First(&student, studentID).Error
	if err != nil {
		return errors.Wrap(err, "PUSH: Error when getting student")
	}
	// Получаем список дней у ученика
	err = pg.Model(&student).Related(&days).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting student='%v' days", student)
	}
	// Гоняем по дням из пакета
	for _, day := range week.Data {
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
				return errors.Wrapf(err, "Error creating newDay='%v'", newDay)
			}
			days = append(days, newDay)
		}
		// Получаем список заданий для дня
		err = pg.Model(&newDay).Related(&tasks).Error
		if err != nil {
			return errors.Wrapf(err, "Error getting newDay='%v' tasks", newDay)
		}
		// Гоняем по заданиям
		for _, task := range day.Lessons {
			// Найдем подходящее задание в БД
			dbTaskFound := false
			for _, dbTask := range tasks {
				if task.AID == dbTask.AID {
					dbTaskFound = true
					newTask = dbTask
					// Сравнить оценки
					if task.Mark != dbTask.Mark {
						// Если оценки не совпали
						important := false
						// Если срезовая оценка или контрольная, она важна
						if task.Type == "В" || task.Type == "К" {
							important = true
						}
						// тип работы
						typ := "что то"
						switch task.Type {
						case "А":
							typ = "практическую работу"
						case "В":
							typ = "срезовую работу"
						case "Д":
							typ = "домашнюю работу"
						case "К":
							typ = "контрольную работу"
						case "С":
							typ = "самостоятельную работу"
						case "Л":
							typ = "лабораторную работу"
						case "П":
							typ = "проект"
						case "Н":
							typ = "диктант"
						case "Р":
							typ = "реферат"
						case "О":
							typ = "ответ на уроке"
						case "Ч":
							typ = "сочинение"
						case "И":
							typ = "изложение"
						case "З":
							typ = "зачёт"
						case "Т":
							typ = "тестирование"
						default:
							typ = "что-то"
						}
						body := "У вас " + dbTask.Mark + " за " + typ + " " + date
						tasksMarks.TasksMarks = append(tasksMarks.TasksMarks, diaryNewTaskMark{Title: task.Name, Body: body, IsImportant: important, IsHomework: false, Type: Mark})
						// сохранить в бд
						dbTask.Mark = task.Mark
						err = pg.Save(&dbTask).Error
						if err != nil {
							return errors.Wrapf(err, "Error saving newTask='%v'", newTask)
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
					return errors.Wrapf(err, "Error creating newTask='%v'", newTask)
				}
				tasks = append(tasks, newTask)
				// Новое задание, запишем в счетчик
				homework := false
				if newTask.Type == "Д" {
					// Если домашняя работа, так же обновим счетчик
					homework = true
				}
				tasksMarks.TasksMarks = append(tasksMarks.TasksMarks, diaryNewTaskMark{Title: task.Name, Body: task.Title, IsImportant: false, IsHomework: homework, Type: Task})
			}
		}
		err = pg.Save(&newDay).Error
		if err != nil {
			return errors.Wrapf(err, "Error saving newDay='%v'", newDay)
		}
	}
	err = pg.Save(&student).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving student='%v'", student)
	}
	return nil
}

// mailNewMessages
type mailNewMessages struct {
	Messages []mailNewMessage
}

// mailNewMessage
type mailNewMessage struct {
	Title string
	Body  string
}

// checkMail
func (p *Push) checkMail(userID uint, mailsList *dt.EmailsList, res *mailNewMessages) error {
	var (
		user       db.User
		newMessage db.MailMessage
		messages   []db.MailMessage
	)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем пользователя по pk userID
	err := pg.First(&user, userID).Error
	if err != nil {
		return errors.Wrap(err, "PUSH: Error when getting user")
	}
	// Получаем сообщения у почты
	err = pg.Model(&user).Related(&messages).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' messages", user)
	}
	// Гоняем по сообщениям из пакета
	for _, post := range mailsList.Record {
		// Найдем подходящее сообщение в БД
		postFound := false
		for _, dbPost := range messages {
			if post.MessageID == dbPost.NetschoolID {
				postFound = true
				newMessage = dbPost
				break
			}
		}
		if !postFound {
			unread := true
			if post.Read == "Y" {
				unread = false
			}
			// Сообщения не существует, надо создать
			newMessage = db.MailMessage{UserID: user.ID, Section: 1, NetschoolID: post.MessageID, Date: post.Sent, Author: post.FromName, Unread: unread, Topic: post.Subj}
			err = pg.Create(&newMessage).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newMessage='%v'", newMessage)
			}
			messages = append(messages, newMessage)
			// Запишем сообщение в структуру
			res.Messages = append(res.Messages, mailNewMessage{Title: newMessage.Author, Body: newMessage.Topic})
		}
	}
	// Сохраним пользователя
	err = pg.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}

// forumNewMessages
type forumNewMessages struct {
	Messages []forumNewMessage
}

// forumNewMessage
type forumNewMessage struct {
	Title    string
	Subtitle string
	Body     string
}

// checkForumTopics
func (p *Push) checkForumTopics(userID uint, themes *dt.ForumThemesList) error {
	var (
		user     db.User
		newTopic db.ForumTopic
		topics   []db.ForumTopic
	)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем пользователя по pk userID
	err := pg.First(&user, userID).Error
	if err != nil {
		return errors.Wrap(err, "PUSH: Error when getting user")
	}
	// Получаем список тем у пользователя
	err = pg.Model(&user).Related(&topics).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' forum topics", user)
	}
	// Гоняем по темам из пакета
	for _, topic := range themes.Posts {
		// Найдем подходящую тему в БД
		topicFound := false
		for _, dbTopic := range topics {
			if topic.ID == dbTopic.NetschoolID {
				topicFound = true
				newTopic = dbTopic
				break
			}
		}
		if !topicFound {
			// Темы не существует, надо создать
			newTopic = db.ForumTopic{UserID: user.ID, NetschoolID: topic.ID, Date: topic.Date, Creator: topic.Creator, Title: topic.Title, Unread: true, Answers: topic.Answers, Posts: []db.ForumPost{}}
			err = pg.Create(&newTopic).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newTopic='%v'", newTopic)
			}
			topics = append(topics, newTopic)
		}
	}
	// Сохраним пользователя
	err = pg.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}

// checkForumMessages
func (p *Push) checkForumMessages(userID uint, themeID int, themeTitle string, topics *dt.ForumThemeMessages, ms *forumNewMessages) error {
	var (
		user       db.User
		topic      db.ForumTopic
		newMessage db.ForumPost
		messages   []db.ForumPost
	)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем пользователя по pk userID
	err := pg.First(&user, userID).Error
	if err != nil {
		return errors.Wrap(err, "PUSH: Error when getting user")
	}
	// Получаем нужную тему у пользователя
	wh := db.ForumTopic{NetschoolID: themeID, UserID: userID}
	err = pg.Where(wh).First(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting forum topic='%v'", wh)
	}
	// Получаем сообщения у темы
	err = pg.Model(&topic).Related(&messages).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting topic='%v' messages", topic)
	}
	// Гоняем по сообщениям из пакета
	for _, post := range topics.Messages {
		// Найдем подходящее сообщение в БД
		postFound := false
		for _, dbPost := range messages {
			if post.Date == dbPost.Date && post.Author == dbPost.Author && post.Message == dbPost.Message {
				postFound = true
				newMessage = dbPost
				break
			}
		}
		if !postFound {
			// Сообщения не существует, надо создать
			newMessage = db.ForumPost{ForumTopicID: topic.ID, Date: post.Date, Author: post.Author, Unread: true, Message: post.Message}
			err = pg.Create(&newMessage).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newMessage='%v'", newMessage)
			}
			messages = append(messages, newMessage)
			// Пополним возвращаемую структуру
			ms.Messages = append(ms.Messages, forumNewMessage{Title: themeTitle, Subtitle: newMessage.Author, Body: newMessage.Message})
		}
	}
	// Сохраним тему
	err = pg.Save(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving topic='%v'", topic)
	}
	return nil
}

type resourcesChanges struct {
	Changes []resourcesChange
}

type resourcesChange struct {
	Title    string
	Subtitle string
	Body     string
}

// checkSchedule проверяет, были ли изменения в расписании
func (p *Push) checkSchedule(studentID uint, week *dt.TimeTable) (bool, error) {
	var (
		student   db.Student
		newDay    db.Day
		days      []db.Day
		newLesson db.Lesson
		lessons   []db.Lesson
	)
	// Флаг изменений
	changed := false
	// shortcut
	pg := p.db.SchoolServerDB
	// Получаем ученика по pk studentID
	err := pg.First(&student, studentID).Error
	if err != nil {
		return false, errors.Wrap(err, "PUSH: Error when getting student")
	}
	// Получаем список дней у ученика
	err = pg.Model(&student).Related(&days).Error
	if err != nil {
		return false, errors.Wrapf(err, "Error getting student='%v' days", student)
	}
	// Гоняем по дням из пакета
	for _, day := range week.Days {
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
			newDay = db.Day{StudentID: student.ID, Date: date, Tasks: []db.Task{}, Lessons: []db.Lesson{}}
			err = pg.Create(&newDay).Error
			if err != nil {
				return false, errors.Wrapf(err, "Error creating newDay='%v'", newDay)
			}
			days = append(days, newDay)
			// Устанавливаем флаг
		}
		// Получаем список уроков для дня
		err = pg.Model(&newDay).Related(&lessons).Error
		if err != nil {
			return false, errors.Wrapf(err, "Error getting newDay='%v' lessons", newDay)
		}
		// Гоняем по урокам
		for _, lesson := range day.Lessons {
			// Найдем подходящий урок в БД
			dbLessonFound := false
			for _, dbLesson := range lessons {
				if lesson.Begin == dbLesson.Begin {
					if lesson.Name != dbLesson.Name || lesson.ClassRoom != dbLesson.Classroom {
						// Произошли изменения
						// по полю Begin сравнивали, они равны, End наверное тоже
						dbLesson.Name = lesson.Name
						dbLesson.Classroom = lesson.ClassRoom
						err = pg.Save(&dbLesson).Error
						if err != nil {
							return false, errors.Wrapf(err, "Error saving updated lesson='%v'", dbLesson)
						}
						// Устанавливаем флаг
						changed = true
					}
					// Если урок нашелся, обновим поля в БД
					dbLessonFound = true
					newLesson = dbLesson
					break
				}
			}
			if !dbLessonFound {
				if len(lessons) == 1 && lessons[0].Begin == "00:00" {
					// Значит, что выходной перестал быть выходным
					// Псевдоудаляем
					err = pg.Delete(&lessons[0]).Error
					if err != nil {
						return false, errors.Wrapf(err, "Error deleting lesson='%v'", lessons[0])
					}
					// Устанавливаем флаг
					changed = true
				}
				// Урока не существует, надо создать
				newLesson = db.Lesson{DayID: newDay.ID, Begin: lesson.Begin, End: lesson.End, Name: lesson.Name, Classroom: lesson.ClassRoom}
				err = pg.Create(&newLesson).Error
				if err != nil {
					return false, errors.Wrapf(err, "Error creating newLesson='%v'", newLesson)
				}
				lessons = append(lessons, newLesson)
			}
		}
		err = pg.Save(&newDay).Error
		if err != nil {
			return false, errors.Wrapf(err, "Error saving newDay='%v'", newDay)
		}
	}
	err = pg.Save(&student).Error
	if err != nil {
		return false, errors.Wrapf(err, "Error saving student='%v'", student)
	}
	return changed, nil
}

// checkResources считает число новых ресурсов у школы
func (p *Push) checkResources(schoolID uint, resources *dt.Resources) (*resourcesChanges, error) {
	var (
		school      db.School
		subs        []db.ResourceSubgroup
		groups      []db.ResourceGroup
		reses       []db.Resource
		newGroup    db.ResourceGroup
		newSubgroup db.ResourceSubgroup
		newRes      db.Resource
		newPart     bool
		change      resourcesChange
	)
	changes := make([]resourcesChange, 0)
	// shortcut
	pg := p.db.SchoolServerDB
	// Получим школу по primary key
	err := pg.First(&school, schoolID).Error
	if err != nil {
		return nil, errors.Wrapf(err, "PUSH: Error when getting school by primary key='%v'", schoolID)
	}
	// Получим все ресурсы школы
	where := db.ResourceGroup{SchoolID: schoolID}
	err = pg.Where(where).Find(&groups).Error
	if err != nil {
		return nil, errors.Wrapf(err, "PUSH: Error when getting resource for school with primary key='%v'", schoolID)
	}
	// Гоним по группам ресурсов
	for _, rGroup := range resources.Data {
		// Если новый раздел
		newPart = false
		// Найдем подходящую группу в БД
		groupFound := false
		for _, g := range groups {
			if g.Title == rGroup.GroupTitle {
				groupFound = true
				newGroup = g
				break
			}
		}
		if !groupFound {
			// Группы не существует, надо создать
			newGroup = db.ResourceGroup{SchoolID: schoolID, Title: rGroup.GroupTitle, Resources: []db.Resource{}, ResourceSubgroups: []db.ResourceSubgroup{}}
			err = pg.Create(&newGroup).Error
			if err != nil {
				return nil, errors.Wrapf(err, "Error creating newGroup='%v'", newGroup)
			}
			// обновить счетчик
			change.Title = "Новый раздел"
			change.Subtitle = ""
			change.Body = rGroup.GroupTitle
			changes = append(changes, change)
			newPart = true
		}
		// Получаем список ресурсов для группы
		err = pg.Model(&newGroup).Related(&reses, "Resources").Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting newGroup='%v' resources", newGroup)
		}
		// Гоняем по ресурсам
		for _, res := range rGroup.Files {
			// Найдем подходящий ресурс в БД
			resFound := false
			for _, r := range reses {
				if res.Name == r.Name {
					resFound = true
					newRes = r
					break
				}
			}
			if !resFound {
				// Ресурса не существует, надо создать
				newRes = db.Resource{Name: res.Name, Link: res.Link}
				err = pg.Create(&newRes).Error
				if err != nil {
					return nil, errors.Wrapf(err, "Error creating newRes='%v'", newRes)
				}
				reses = append(reses, newRes)
				// Новый файл в существующем разделе, обновим счетчик
				if !newPart {
					change.Title = "Новый файл"
					change.Subtitle = rGroup.GroupTitle
					change.Body = res.Name
					changes = append(changes, change)
				}
			}
		}
		// Сохраним группу и ее ресурсы и подгруппы
		newGroup.Resources = reses
		err = pg.Save(&newGroup).Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error saving newGroup='%v'", newGroup)
		}
		// Получаем список подгрупп этой группы
		err = pg.Model(&newGroup).Related(&subs, "ResourceSubgroups").Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error getting newGroup='%v' subgroups", newGroup)
		}
		// Гоняем по подгруппам
		for _, sub := range rGroup.Subgroups {
			// Найдем подходящую подгруппу в БД
			groupFound := false
			for _, g := range subs {
				if g.Title == sub.SubgroupTitle {
					groupFound = true
					newSubgroup = g
					break
				}
			}
			if !groupFound {
				// Подгруппы не существует, надо создать
				newSubgroup = db.ResourceSubgroup{ResourceGroupID: newGroup.ID, Title: sub.SubgroupTitle, Resources: []db.Resource{}}
				err = pg.Create(&newSubgroup).Error
				if err != nil {
					return nil, errors.Wrapf(err, "Error creating newSubgroup='%v'", newSubgroup)
				}
			}
			// Получаем список ресурсов для подгруппы
			err = pg.Model(&newSubgroup).Related(&reses, "Resources").Error
			if err != nil {
				return nil, errors.Wrapf(err, "Error getting newSubgroup='%v' resources", newGroup)
			}
			// Гоняем по ресурсам
			for _, res := range sub.Files {
				// Найдем подходящий ресурс в БД
				resFound := false
				for _, r := range reses {
					if res.Name == r.Name {
						resFound = true
						newRes = r
						break
					}
				}
				if !resFound {
					// Ресурса не существует, надо создать
					newRes = db.Resource{Name: res.Name, Link: res.Link}
					err = pg.Create(&newRes).Error
					if err != nil {
						return nil, errors.Wrapf(err, "Error creating newRes='%v'", newRes)
					}
					reses = append(reses, newRes)
					// Новый файл в существующем подразделе, обновим счетчик
					if !newPart {
						change.Title = "Новый файл"
						change.Subtitle = rGroup.GroupTitle
						change.Body = res.Name
						changes = append(changes, change)
					}
				}
			}
			// Сохраним подгруппу и ее ресурсы
			newSubgroup.Resources = reses
			err = pg.Save(&newSubgroup).Error
			if err != nil {
				return nil, errors.Wrapf(err, "Error saving newSubgroup='%v'", newSubgroup)
			}
		}
		// Сохраним группу и ее ресурсы и подгруппы
		err = pg.Save(&newGroup).Error
		if err != nil {
			return nil, errors.Wrapf(err, "Error saving newGroup='%v'", newGroup)
		}
	}
	return &resourcesChanges{Changes: changes}, nil
}

type GorushRequest struct {
	Notifications []Notification `json:"notifications"`
}

type Alert struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Body     string `json:"body,omitempty"`
}

type Notification struct {
	Tokens         []string `json:"tokens,omitempty"`
	Platform       int      `json:"platform,omitempty"`
	Badge          int      `json:"badge,omitempty"`
	Category       string   `json:"category,omitempty"`
	MutableContent bool     `json:"mutableContent,omitempty"`
	Topic          string   `json:"topic,omitempty"`
	Alert          Alert    `json:"alert,omitempty"`
	Message        string   `json:"message,omitempty"`
	Sound          string   `json:"sound,omitempty"`
}

// send посылает push-уведомление по web api gorush
func (p *Push) send(systemType int, token, category, title, subtitle, body, message string) error {
	var notifications []Notification
	notifications = make([]Notification, 1)
	notifications[0] = Notification{
		Tokens:         []string{token},
		Platform:       systemType,
		Badge:          1,
		Category:       category,
		MutableContent: true,
		Topic:          p.AppTopic,
		Message:        message,
		Sound:          "default",
		Alert: Alert{
			Title:    title,
			Subtitle: subtitle,
			Body:     body,
		},
	}
	req := GorushRequest{Notifications: notifications}
	byt, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "PUSH: Error marshalling norification")
	}
	resp, err := http.Post(p.GorushAddress, "application/json", bytes.NewBuffer(byt))
	if err != nil {
		return errors.Wrap(err, "PUSH: Error sending web api gorush request")
	}
	defer resp.Body.Close()
	p.logger.Info("PUSH: Got response from gorush", "Response", resp)
	return nil
}
