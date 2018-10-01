/*
Package push содержит объявления функций, посылающих пуши.
*/
package push

import (
	"strconv"
	"time"

	"context"

	"github.com/appleboy/gorush/rpc/proto"
	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	api "github.com/masyagin1998/SchoolServer/libtelco/rest-api"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	db "github.com/masyagin1998/SchoolServer/libtelco/sql-db"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Push struct содержит конфигурацию пушей.
type Push struct {
	api     *api.RestAPI
	db      *db.Database
	logger  *log.Logger
	stopped bool
	period  time.Duration
	client  *proto.GorushClient
}

// NewPush создает структуру пушей и возвращает указатель на неё.
func NewPush(restapi *api.RestAPI, logger *log.Logger) *Push {
	return &Push{
		api:     restapi,
		db:      restapi.Db,
		logger:  logger,
		stopped: true,
		period:  time.Second * 15,
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
			go p.handlePushes()
		}
	}
}

// Stop останавливает рассылку.
func (p *Push) Stop() {
	p.logger.Info("PUSH: Stoppped")
	p.stopped = true
}

//
func (p *Push) handlePushes() {
	var (
		users   []db.User
		devices []db.Device
		school  db.School
	)
	p.logger.Info("PUSH: Sending push notifications")
	// shortcut
	pg := p.db.SchoolServerDB
	// Достанем всех пользователей
	err := pg.Find(&users).Error
	if err != nil {
		p.logger.Error("PUSH: Error when getting users list", "Error", err)
		return
	}
	// Отображение schoolID в число новых для этой школы
	nResources := make(map[uint]*resourcesChanges)
	// Текущее время
	now := time.Now()
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
			p.logger.Info("kek", "changes", rChanges)
			if rChanges != nil {
				alert := proto.Alert{
					Title:    "Test Title",
					Body:     "Test Alert Body",
					Subtitle: "Test Alert Sub Title",
					LocKey:   "Test loc key",
					LocArgs:  []string{"test", "test"},
				}
				for _, v := range rChanges.Changes {
					alert.Title = v.Title
					alert.Body = v.Body
					alert.Subtitle = v.Subtitle
					if v.Subtitle == "" {
						// это группа
						err = p.sendGRPC("", "resources_new_file_group", &alert, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "alert", alert, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					} else {
						// это файл
						err = p.sendGRPC("", "resources_new_file", &alert, dev.SystemType, dev.Token)
						if err != nil {
							p.logger.Error("PUSH: Error when sending push to client", "Error", err, "alert", alert, "Platform Type", dev.SystemType, "Token", dev.Token)
							return
						}
					}
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

type resourcesChanges struct {
	Changes []resourcesChange
}

type resourcesChange struct {
	Title    string
	Subtitle string
	Body     string
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
	return &resourcesChanges{
		Changes: changes,
	}, nil
}

type gorushRequest struct {
	Notifications []notification `json:"notifications"`
}

type notification struct {
	Tokens   []string `json:"tokens"`
	Platform int      `json:"platform"`
	Message  string   `json:"message"`
}

func (p *Push) sendGRPC(msg, category string, alert *proto.Alert, platformType int, token string) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := proto.NewGorushClient(conn)
	_, err = c.Send(context.Background(), &proto.NotificationRequest{
		Platform: int32(platformType),
		Tokens:   []string{token},
		//Message:        "",
		Badge:          1,
		Category:       category,
		Sound:          "default",
		MutableContent: true,
		Alert:          alert,
		Topic:          "kir4567.NetSchoolApp",
	})
	if err != nil {
		return err
	}
	return nil
}
