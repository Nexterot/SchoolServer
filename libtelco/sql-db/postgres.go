/*
Package db содержит необходимое API для работы с базой данных PostgreSQL с использованием библиотеки gorm
*/
package db

import (
	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

// Database struct представляет абстрактную структуру базы данных
type Database struct {
	SchoolServerDB *gorm.DB
	Logger         *log.Logger
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
	// Если таблицы с устройствами не существует, создадим её
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
	// Если таблицы с объявлениями не существует, создадим её
	if !sdb.HasTable(&Post{}) {
		// Post
		logger.Info("DB: Creating 'posts' table")
		err = sdb.CreateTable(&Post{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'posts' table")
	} else {
		logger.Info("DB: Table 'posts' exists")
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
	// Если таблицы с уроками не существует, создадим её
	if !sdb.HasTable(&Lesson{}) {
		// Lesson
		logger.Info("DB: Creating 'lessons' table")
		err = sdb.CreateTable(&Lesson{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'lessons' table")
	} else {
		logger.Info("DB: Table 'lessons' exists")
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

// Close закрывает БД
func (db *Database) Close() error {
	return db.SchoolServerDB.Close()
}
