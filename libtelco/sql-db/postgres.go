/*
Package db содержит необходимое API для работы с базой данных PostgreSQL с использованием библиотеки gorm
*/
package db

import (
	"fmt"
	"strconv"
	"time"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
	"github.com/pkg/errors"
)

// Статусы заданий
const (
	_ = iota
	StatusTaskNew
	StatusTaskSeen
	StatusTaskDone
)

// Типы систем устройств
const (
	_ = iota
	IOS
	Android
)

// Типы оповещений об оценках
const (
	_ = iota
	MarksNotificationAll
	MarksNotificationImportant
	MarksNotificationDisabled
)

// Типы оповещений о заданиях
const (
	_ = iota
	TasksNotificationAll
	TasksNotificationHome
	TasksNotificationDisabled
)

// Database struct представляет абстрактную структуру базы данных
type Database struct {
	SchoolServerDB *gorm.DB
	Logger         *log.Logger
}

// School struct представляет структуру записи школы
type School struct {
	gorm.Model
	Name       string `sql:"size:255;unique"`
	Type       int    // тип netschool'а
	Initials   string `sql:"size:255"`
	Address    string `sql:"size:255;unique"`
	Permission bool   `sql:"DEFAULT:true"`
	Users      []User // has-many relation
}

// User struct представляет структуру записи пользователя
type User struct {
	gorm.Model
	SchoolID     uint // parent id
	UID          int
	Login        string        `sql:"size:255;index"`
	Password     string        `sql:"size:255"`
	Permission   bool          `sql:"DEFAULT:false"`
	FirstName    string        `sql:"size:255"`
	LastName     string        `sql:"size:255"`
	TrueLogin    string        `sql:"size:255"`
	Role         string        `sql:"size:255"`
	Year         string        `sql:"size:255"`
	Students     []Student     // has-many relation
	Devices      []Device      // has-many relation
	ForumTopics  []ForumTopic  // has-many relation
	MailMessages []MailMessage //  has-many relation
}

// Device struct представляет структуру устройства, которое будет получать
// push-уведомления
type Device struct {
	gorm.Model
	UserID                uint       // parent id
	SystemType            int        // android/ios
	Token                 string     // unique device token (FCM)
	DoNotDisturbUntil     *time.Time // для push
	MarksNotification     int        `sql:"DEFAULT:1"`
	TasksNotification     int        `sql:"DEFAULT:1"`
	ReportsNotification   bool       `sql:"DEFAULT:true"`
	ScheduleNotification  bool       `sql:"DEFAULT:true"`
	MailNotification      bool       `sql:"DEFAULT:true"`
	ForumNotification     bool       `sql:"DEFAULT:true"`
	ResourcesNotification bool       `sql:"DEFAULT:true"`
}

// Student struct представляет структуру ученика
type Student struct {
	gorm.Model
	UserID      uint   // parent id
	Name        string `sql:"size:255"`
	NetSchoolID int    // id ученика в системе NetSchool
	ClassID     string // id класса, в котором учится ученик
	Days        []Day  // has-many relation
}

// Day struct представляет структуру дня с дз
type Day struct {
	gorm.Model
	StudentID uint   // parent id
	Date      string `sql:"size:255"`
	Tasks     []Task // has-many relation
}

// Task struct представляет структуру задания
type Task struct {
	gorm.Model
	DayID  uint // parent id
	CID    int
	AID    int
	TP     int
	Status int // новое\просмотренное\выполненное
	InTime bool
	Name   string // название предмета
	Title  string // тема
	Type   string
	Mark   string
	Weight string
	Author string
}

// ForumTopic struct представляет структуру темы на форуме
type ForumTopic struct {
	gorm.Model
	UserID      uint // parent id
	NetschoolID int
	Date        string
	Creator     string
	Title       string
	Unread      bool
	Answers     int
	Posts       []ForumPost // has-many relation
}

// ForumPost struct представляет структуру сообщения в теме на форуме
type ForumPost struct {
	gorm.Model
	ForumTopicID uint // parent id
	Date         string
	Author       string
	Message      string
	Unread       bool
}

// MailMessage struct представляет структуру сообщения на почте
type MailMessage struct {
	gorm.Model
	UserID      uint //  parent id
	NetschoolID int
	Date        string
	Author      string
	Topic       string
	Section     int
	Unread      bool
}

// Resource struct представляет структуру школьного ресурса
type Resource struct {
	gorm.Model
	OwnerID   uint   // parent id
	OwnerType string // parent polymorhic type
	Name      string
	Link      string
}

// ResourceGroup struct представляет структуру группы ресурсов
type ResourceGroup struct {
	gorm.Model
	SchoolID          uint // belongs-to relation
	Title             string
	Resources         []Resource         `gorm:"polymorphic:Owner;"`
	ResourceSubgroups []ResourceSubgroup // has-many relation
}

// ResourceSubgroup struct представляет структуру подгруппы ресурсов
type ResourceSubgroup struct {
	gorm.Model
	ResourceGroupID uint
	Title           string
	Resources       []Resource `gorm:"polymorphic:Owner;"`
}

// NewDatabase создает Database и возвращает указатель на неё
func NewDatabase(logger *log.Logger, config *cp.Config) (*Database, error) {
	// Подключение к базе данных
	conf := "host=" + config.Postgres.Host + " port=" + config.Postgres.Port +
		" user=" + config.Postgres.User + " password=" + config.Postgres.Password +
		" dbname=" + config.Postgres.DBname + " sslmode=" + config.Postgres.SSLmode
	sdb, err := gorm.Open("postgres", conf)
	if err != nil {
		return nil, err
	}
	// Если таблицы с пользователями не существует, создадим её
	if !sdb.HasTable(&User{}) {
		// User
		logger.Info("DB: Table 'users' doesn't exist! Creating new one")
		err := sdb.CreateTable(&User{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'users' table")
	} else {
		logger.Info("DB: Table 'users' exists")
	}
	// Если таблицы с устрйоствами не существует, создадим её
	if !sdb.HasTable(&Device{}) {
		// Device
		logger.Info("DB: Creating 'device' table")
		err = sdb.CreateTable(&Device{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'device' table")
	} else {
		logger.Info("DB: Table 'devices' exists")
	}
	// Если таблицы со студентами не существует, создадим её
	if !sdb.HasTable(&Student{}) {
		// Student
		logger.Info("DB: Creating 'students' table")
		err = sdb.CreateTable(&Student{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'students' table")
	} else {
		logger.Info("DB: Table 'students' exists")
	}
	// Если таблицы со днями не существует, создадим её
	if !sdb.HasTable(&Day{}) {
		// Day
		logger.Info("DB: Creating 'days' table")
		err = sdb.CreateTable(&Day{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'days' table")
	} else {
		logger.Info("DB: Table 'days' exists")
	}
	// Если таблицы с заданиями не существует, создадим её
	if !sdb.HasTable(&Task{}) {
		// Task
		logger.Info("DB: Creating 'tasks' table")
		err = sdb.CreateTable(&Task{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'tasks' table")
	} else {
		logger.Info("DB: Table 'tasks' exists")
	}
	// Если таблицы с ресурсами не существует, создадим её
	if !sdb.HasTable(&Resource{}) {
		// Resource
		logger.Info("DB: Creating 'resources' table")
		err = sdb.CreateTable(&Resource{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'resources' table")
	} else {
		logger.Info("DB: Table 'resources' exists")
	}
	// Если таблицы с группой ресурсов не существует, создадим её
	if !sdb.HasTable(&ResourceGroup{}) {
		// ResourceGroup
		logger.Info("DB: Creating 'resource_groups' table")
		err = sdb.CreateTable(&ResourceGroup{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'resource_groups' table")
	} else {
		logger.Info("DB: Table 'resource_groups' exists")
	}
	// Если таблицы с подгруппой ресурсов не существует, создадим её
	if !sdb.HasTable(&ResourceSubgroup{}) {
		// ResourceSubgroup
		logger.Info("DB: Creating 'resource_subgroups' table")
		err = sdb.CreateTable(&ResourceSubgroup{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'resource_subgroups' table")
	} else {
		logger.Info("DB: Table 'resource_subgroups' exists")
	}
	// Если таблицы с темами форума не существует, создадим её
	if !sdb.HasTable(&ForumTopic{}) {
		// ForumTopic
		logger.Info("DB: Creating 'forum_topics' table")
		err = sdb.CreateTable(&ForumTopic{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'forum_topics' table")
	} else {
		logger.Info("DB: Table 'forum_topics' exists")
	}
	// Если таблицы с сообщениями форума не существует, создадим её
	if !sdb.HasTable(&ForumPost{}) {
		// ForumPost
		logger.Info("DB: Creating 'forum_posts' table")
		err = sdb.CreateTable(&ForumPost{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'forum_posts' table")
	} else {
		logger.Info("DB: Table 'forum_posts' exists")
	}
	// Если таблицы с сообщениями почты не существует, создадим её
	if !sdb.HasTable(&MailMessage{}) {
		// ForumPost
		logger.Info("DB: Creating 'mail_messages' table")
		err = sdb.CreateTable(&MailMessage{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'mail_messages' table")
	} else {
		logger.Info("DB: Table 'mail_messages' exists")
	}
	// Если таблицы со школами не существует, создадим её
	if !sdb.HasTable(&School{}) {
		logger.Info("DB: 'schools' table doesn't exist! Creating new one")
		err := sdb.CreateTable(&School{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'schools' table")
	} else {
		logger.Info("DB: Table 'schools' exists")
	}
	var count int
	err = sdb.Table("schools").Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count == 0 {
		logger.Info("DB: Table 'schools' empty. Adding our schools")
		// Добавим записи поддерживаемых школ
		school1 := School{
			Address: "62.117.74.43", Name: "Европейская гимназия", Initials: "ЕГ", Type: 1,
		}
		err = sdb.Save(&school1).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully added our schools")
	}
	return &Database{SchoolServerDB: sdb, Logger: logger}, nil
}

// UpdateUser обновляет данные о пользователе
func (db *Database) UpdateUser(login string, passkey string, schoolID int, token string, systemType int, childrenMap map[string]dt.Student, profile *dt.Profile) error {
	var (
		school  School
		student Student
		user    User
		device  Device
	)
	// Получаем запись школы
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting school with id='%v'", schoolID)
	}
	// Получаем запись пользователя
	where := User{Login: login, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Unscoped().Where(where).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			// Пользователь не найден, создадим
			students := make([]Student, len(childrenMap))
			i := 0
			for childName, childInfo := range childrenMap {
				sid, errInner := strconv.Atoi(childInfo.SID)
				if errInner != nil {
					return errors.Wrapf(errInner, "Error converting SID='%s' from string to int", childInfo.SID)
				}
				student = Student{Name: childName, NetSchoolID: sid, Days: []Day{}, ClassID: childInfo.CLID}
				errInner = db.SchoolServerDB.Create(&student).Error
				if errInner != nil {
					return errors.Wrapf(errInner, "Error creating student='%v' for user='%v'", student, user)
				}
				students[i] = student
				i++
			}
			// Создать список девайсов
			devices := make([]Device, 1)
			dev := Device{SystemType: systemType, Token: token}
			// Создать девайс
			errInner := db.SchoolServerDB.Create(&dev).Error
			devices[0] = dev
			if errInner != nil {
				return errors.Wrapf(errInner, "Error creating device='%v' for user='%v'", dev, user)
			}
			user := User{
				SchoolID:     uint(schoolID),
				Login:        login,
				Password:     passkey,
				Students:     students,
				Devices:      devices,
				Role:         profile.Role,
				LastName:     profile.Surname,
				FirstName:    profile.Name,
				Year:         profile.Schoolyear,
				TrueLogin:    profile.Username,
				ForumTopics:  []ForumTopic{},
				MailMessages: []MailMessage{},
			}
			// Записываем профиль
			user.UID, err = strconv.Atoi(profile.UID)
			if err != nil {
				return errors.Wrapf(err, "Error converting string to int: %v'", profile.UID)
			}
			errInner = db.SchoolServerDB.Create(&user).Error
			if errInner != nil {
				return errors.Wrapf(errInner, "Error creating user='%v'", user)
			}
			db.Logger.Info("User created", "User", user)
			return nil
		}
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Пользователь найден, обновим
	user.Password = passkey
	user.SchoolID = uint(schoolID)
	// Если девайс новый, добавить
	wh := Device{UserID: user.ID, Token: token, SystemType: systemType}
	err = db.SchoolServerDB.Unscoped().FirstOrCreate(&device, wh).Error
	if err != nil {
		return errors.Wrapf(err, "Error query firstorcreate device='%v'", wh)
	}
	// Если девайс был удален
	if device.DeletedAt != nil {
		device.DeletedAt = nil
		err = db.SchoolServerDB.Unscoped().Save(&device).Error
		if err != nil {
			return errors.Wrapf(err, "Error saving updated deleted device='%v'", device)
		}
	}
	// Если пользователь был псевдоудален
	if user.DeletedAt != nil {
		user.DeletedAt = nil
		err = db.SchoolServerDB.Unscoped().Save(&user).Error
		if err != nil {
			return errors.Wrapf(err, "Error saving updated deleted user='%v'", user)
		}
	} else {
		err = db.SchoolServerDB.Save(&user).Error
		if err != nil {
			return errors.Wrapf(err, "Error saving updated user='%v'", user)
		}
	}
	return nil
}

// UpdatePushTime обновляет время, до которого не беспокоить пользователя push-уведомлениями
func (db *Database) UpdatePushTime(userName string, schoolID int, token string, systemType int, minutes int) error {
	var (
		user   User
		school School
		device Device
	)
	// Получаем школу по id
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return errors.Wrapf(err, "Error query school by id='%v'", schoolID)
	}
	// Получаем пользователя по школе и логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем устройство по token и UserID
	wh := Device{UserID: user.ID, Token: token, SystemType: systemType}
	err = db.SchoolServerDB.Where(wh).First(&device).Error
	if err != nil {
		return errors.Wrapf(err, "Error query device='%v'", wh)
	}
	// Обновим поле
	t := time.Now().Add(time.Minute * time.Duration(minutes))
	device.DoNotDisturbUntil = &t
	// Сохраним
	err = db.SchoolServerDB.Save(&device).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving device='%v'", user)
	}
	return nil
}

// PseudoDeleteUser псевдоудаляет пользователя
func (db *Database) PseudoDeleteUser(userName string, schoolID int) error {
	var (
		user   User
		school School
	)
	// Получаем школу по id
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return errors.Wrapf(err, "Error query school by id='%v'", schoolID)
	}
	// Получаем пользователя по школе и логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Псевдоудаляем
	return db.SchoolServerDB.Delete(&user).Error
}

// PseudoDeleteDevice псевдоудаляет девайс пользователя
func (db *Database) PseudoDeleteDevice(userName string, schoolID int, token string, systemType int) error {
	var (
		user   User
		school School
		device Device
	)
	// Получаем школу по id
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return errors.Wrapf(err, "Error query school by id='%v'", schoolID)
	}
	// Получаем пользователя по школе и логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем устройство по token и UserID
	wh := Device{UserID: user.ID, Token: token, SystemType: systemType}
	err = db.SchoolServerDB.Where(wh).First(&device).Error
	if err != nil {
		return errors.Wrapf(err, "Error query device='%v'", wh)
	}
	// Псевдоудаляем
	return db.SchoolServerDB.Delete(&device).Error
}

// GetUserProfile возвращает профиль пользователя
func (db *Database) GetUserProfile(userName string, schoolID int) (*dt.Profile, error) {
	var (
		user   User
		school School
	)
	// Получаем школу по id
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query school by id='%v'", schoolID)
	}
	// Получаем пользователя по школе и логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Формируем структуру
	profile := dt.Profile{Role: user.Role, Surname: user.LastName, Name: user.FirstName, Schoolyear: user.Year, UID: strconv.Itoa(user.UID), Username: user.TrueLogin}
	return &profile, nil
}

// GetUserAuthData возвращает данные для повторной удаленной авторизации пользователя
func (db *Database) GetUserAuthData(userName string, schoolID int) (*cp.School, error) {
	var (
		user   User
		school School
	)
	// Получаем школу по id
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query school by id='%v'", schoolID)
	}
	// Получаем пользователя по школе и логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query user='%v'", where)
	}
	return &cp.School{Link: school.Address, Login: userName, Password: user.Password, Type: 1}, nil
}

// GetUserPermission проверяет разрешение пользователя на работу с сервисом
func (db *Database) GetUserPermission(userName string, schoolID int) (bool, error) {
	var user User
	// Получаем пользователя по логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return false, errors.Wrapf(err, "Error query user='%v'", where)
	}
	return user.Permission, nil
}

// UpdateTasksStatuses добавляет в БД несуществующие ДЗ и обновляет статусы
// заданий из пакета WeekSchoolMarks
func (db *Database) UpdateTasksStatuses(userName string, schoolID int, studentID int, week *dt.WeekSchoolMarks) error {
	var (
		student  Student
		students []Student
		user     User
		newDay   Day
		days     []Day
		newTask  Task
		tasks    []Task
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем ученика по studentID
	err = db.SchoolServerDB.Model(&user).Related(&students).Error
	if err != nil {
		return errors.Wrapf(err, "Error query students for user='%v'", user)
	}
	studentFound := false
	for _, stud := range students {
		if stud.NetSchoolID == studentID {
			studentFound = true
			student = stud
			break
		}
	}
	if !studentFound {
		return errors.Wrapf(fmt.Errorf("record not found"), "No student with id='%v' found for userName='%s'", studentID, userName)
	}
	// Получаем список дней у ученика
	err = db.SchoolServerDB.Model(&student).Related(&days).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting student='%v' days", student)
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
			newDay = Day{StudentID: student.ID, Date: date, Tasks: []Task{}}
			err = db.SchoolServerDB.Create(&newDay).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newDay='%v'", newDay)
			}
			days = append(days, newDay)
		}
		// Получаем список заданий для дня
		err = db.SchoolServerDB.Model(&newDay).Related(&tasks).Error
		if err != nil {
			return errors.Wrapf(err, "Error getting newDay='%v' tasks", newDay)
		}
		// Гоняем по заданиям
		for taskNum, task := range day.Lessons {
			// Найдем подходящее задание в БД
			dbTaskFound := false
			for _, dbTask := range tasks {
				if task.AID == dbTask.AID {
					dbTaskFound = true
					newTask = dbTask
					break
				}
			}
			if !dbTaskFound {
				// Задания не существует, надо создать
				newTask = Task{DayID: newDay.ID, CID: task.CID, AID: task.AID, Status: StatusTaskNew, TP: task.TP, InTime: task.InTime, Name: task.Name,
					Title: task.Title, Type: task.Type, Mark: task.Mark, Weight: task.Weight, Author: task.Author}
				err = db.SchoolServerDB.Create(&newTask).Error
				if err != nil {
					return errors.Wrapf(err, "Error creating newTask='%v'", newTask)
				}
				tasks = append(tasks, newTask)
			}
			// Присвоить статусу таска из пакета статус таска из БД
			week.Data[dayNum].Lessons[taskNum].Status = newTask.Status
		}
		err = db.SchoolServerDB.Save(&newDay).Error
		if err != nil {
			return errors.Wrapf(err, "Error saving newDay='%v'", newDay)
		}
	}
	err = db.SchoolServerDB.Save(&student).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving student='%v'", student)
	}
	return nil
}

// UpdateTopicsStatuses добавляет в БД несуществующие темы форума и обновляет статусы
func (db *Database) UpdateTopicsStatuses(userName string, schoolID int, themes *dt.ForumThemesList) error {
	var (
		user     User
		newTopic ForumTopic
		topics   []ForumTopic
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем список тем у пользователя
	err = db.SchoolServerDB.Model(&user).Related(&topics).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' forum topics", user)
	}
	// Гоняем по темам из пакета
	for topicNum, topic := range themes.Posts {
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
			newTopic = ForumTopic{UserID: user.ID, NetschoolID: topic.ID, Date: topic.Date, Creator: topic.Creator, Title: topic.Title, Unread: true, Answers: topic.Answers, Posts: []ForumPost{}}
			err = db.SchoolServerDB.Create(&newTopic).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newTopic='%v'", newTopic)
			}
			topics = append(topics, newTopic)
		}
		// Присвоить статусу темы из пакета статус темы из БД
		themes.Posts[topicNum].Unread = newTopic.Unread
	}
	// Сохраним пользователя
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}

// UpdatePostsStatuses добавляет в БД несуществующие сообщения форума и обновляет статусы
func (db *Database) UpdatePostsStatuses(userName string, schoolID int, themeID int, topics *dt.ForumThemeMessages) error {
	var (
		user       User
		topic      ForumTopic
		newMessage ForumPost
		messages   []ForumPost
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем нужную тему у пользователя
	wh := ForumTopic{NetschoolID: themeID}
	err = db.SchoolServerDB.Where(wh).First(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting forum topic='%v'", wh)
	}
	// Помечаем тему как прочитанную
	topic.Unread = false
	// Получаем сообщения у темы
	err = db.SchoolServerDB.Model(&topic).Related(&messages).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting topic='%v' messages", topic)
	}
	// Гоняем по сообщениям из пакета
	for postNum, post := range topics.Messages {
		// Найдем подходящее сообщение в БД
		postFound := false
		for _, dbPost := range messages {
			if post.Date == dbPost.Date && post.Author == dbPost.Author && post.Message == dbPost.Message {
				postFound = true
				newMessage = dbPost
				break
				// В этом случае гарантируется, что сообщение уже было прочитано
			}
		}
		if !postFound {
			// Сообщения не существует, надо создать
			newMessage = ForumPost{ForumTopicID: topic.ID, Date: post.Date, Author: post.Author, Unread: false, Message: post.Message}
			err = db.SchoolServerDB.Create(&newMessage).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newMessage='%v'", newMessage)
			}
			messages = append(messages, newMessage)
			// Присвоить статусу сообщения из пакет из БД "не прочитано"
			topics.Messages[postNum].Unread = true
		}
	}
	// Сохраним тему
	err = db.SchoolServerDB.Save(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving topic='%v'", topic)
	}
	return nil
}

// Letter используется в getMailResponse
type Letter struct {
	Date   string `json:"date"`
	ID     int    `json:"id"`
	Author string `json:"author"`
	Topic  string `json:"topic"`
	Unread bool   `json:"unread"`
}

// getMailResponse используется в GetMailHandler
type getMailResponse struct {
	Letters []Letter `json:"letters"`
}

// UpdateMailStatuses добавляет в БД несуществующие сообщения почты и обновляет статусы
func (db *Database) UpdateMailStatuses(userName string, schoolID int, section int, emailsList *dt.EmailsList) (*getMailResponse, error) {
	var (
		user       User
		newMessage MailMessage
		messages   []MailMessage
	)
	letters := make([]Letter, 0)
	// Сформировать ответ по протоколу
	for _, record := range emailsList.Record {
		unread := true
		if record.Read == "Y" {
			unread = false
		}
		letter := Letter{Date: record.Sent, ID: record.MessageID, Author: record.FromName, Topic: record.Subj, Unread: unread}
		letters = append(letters, letter)
	}
	response := getMailResponse{Letters: letters}
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем сообщения у почты
	err = db.SchoolServerDB.Model(&user).Related(&messages).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting user='%v' messages", user)
	}
	// Гоняем по сообщениям из пакета
	for postNum, post := range response.Letters {
		// Найдем подходящее сообщение в БД
		postFound := false
		for _, dbPost := range messages {
			if post.ID == dbPost.NetschoolID {
				postFound = true
				newMessage = dbPost
				break
				// В этом случае гарантируется, что сообщение уже было прочитано
			}
		}
		if !postFound {
			// Сообщения не существует, надо создать
			newMessage = MailMessage{UserID: user.ID, Section: section, NetschoolID: post.ID, Date: post.Date, Author: post.Author, Unread: post.Unread, Topic: post.Topic}
			err = db.SchoolServerDB.Create(&newMessage).Error
			if err != nil {
				return nil, errors.Wrapf(err, "Error creating newMessage='%v'", newMessage)
			}
			messages = append(messages, newMessage)
			// Присвоить статусу сообщения из пакет из БД "не прочитано"
			response.Letters[postNum].Unread = true
		}
	}
	// Сохраним пользователя
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return &response, nil
}

// MarkMailRead меняет статус письма на прочитанное
func (db *Database) MarkMailRead(userName string, schoolID int, section int, ID int) error {
	var (
		user User
		mail MailMessage
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем нужное письмо
	w := MailMessage{NetschoolID: ID, UserID: user.ID, Section: section}
	err = db.SchoolServerDB.Where(w).First(&mail).Error
	if err != nil {
		return errors.Wrapf(err, "Error query mail: '%v'", w)
	}
	// Обновим статус
	mail.Unread = false
	// Сохраним письмо
	err = db.SchoolServerDB.Save(&mail).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving updated mail='%v'", mail)
	}
	return nil
}

// TaskMarkDone меняет статус задания на "Выполненное"
func (db *Database) TaskMarkDone(userName string, schoolID int, AID, CID, TP int) error {
	var (
		tasks   []Task
		day     Day
		student Student
		user    User
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем таски с таким taskID
	w := Task{AID: AID, CID: CID, TP: TP}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		return errors.Wrapf(err, "Error query tasks: '%v'", w)
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query day with id='%v'", t.DayID)
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query student with id='%v'", day.StudentID)
		}
		// Если совпал id пользователя - поменять статус, сохранить и закончить цикл
		if user.ID == student.UserID {
			t.Status = StatusTaskDone
			err = db.SchoolServerDB.Save(&t).Error
			if err != nil {
				return errors.Wrapf(err, "Error saving updated task='%v'", t)
			}
			return nil
		}
	}
	// Таск не найден
	return fmt.Errorf("record not found")
}

// TaskMarkUndone меняет статус задания на "Просмотренное"
func (db *Database) TaskMarkUndone(userName string, schoolID int, AID, CID, TP int) error {
	var (
		tasks   []Task
		day     Day
		student Student
		user    User
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем таски с таким taskID
	w := Task{AID: AID, CID: CID, TP: TP}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		return errors.Wrapf(err, "Error query tasks: '%v'", w)
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query day with id='%v'", t.DayID)
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query student with id='%v'", day.StudentID)
		}
		// Если совпал id пользователя - поменять статус, сохранить и закончить цикл
		if user.ID == student.UserID {
			t.Status = StatusTaskSeen
			err = db.SchoolServerDB.Save(&t).Error
			if err != nil {
				return errors.Wrapf(err, "Error saving updated task='%v'", t)
			}
			return nil
		}
	}
	// Таск не найден
	return fmt.Errorf("record not found")
}

// GetStudentClassID получает classID ученика
func (db *Database) GetStudentClassID(userName string, schoolID int, studentID int) (string, error) {
	var (
		student Student
		user    User
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return "", errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем ученика
	wh := Student{NetSchoolID: studentID, UserID: user.ID}
	err = db.SchoolServerDB.Where(wh).First(&student).Error
	if err != nil {
		return "", errors.Wrapf(err, "Error query student='%v'", wh)
	}
	// Вернуть classID
	return student.ClassID, nil
}

// GetUserUID получает UserUID пользователя
func (db *Database) GetUserUID(userName string, schoolID int) (string, error) {
	var (
		user User
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return "", errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Вернуть UserID
	return strconv.Itoa(user.UID), nil
}

// GetSchoolPermission проверяет разрешение школы на работу с сервисом
func (db *Database) GetSchoolPermission(id int) (bool, error) {
	var school School
	// Получаем школу по primary key
	err := db.SchoolServerDB.First(&school, id).Error
	if err != nil {
		return false, errors.Wrapf(err, "Error query school with id='%v'", id)
	}
	return school.Permission, nil
}

// GetSchools возвращает информацию о всех поддерживаемых школах
func (db *Database) GetSchools() ([]School, error) {
	var schools []School
	err := db.SchoolServerDB.Find(&schools).Error
	return schools, err
}

// CheckPassword сверяет пароль с хранящимся в БД
func (db *Database) CheckPassword(userName string, schoolID int, pass string) (bool, error) {
	var user User
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return false, errors.Wrapf(err, "Error query user='%v'", where)
	}
	return user.Password == pass, nil
}

// Close закрывает БД
func (db *Database) Close() error {
	return db.SchoolServerDB.Close()
}
