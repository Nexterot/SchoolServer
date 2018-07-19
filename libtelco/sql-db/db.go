/*
Package db содержит необходимое API для работы с базой данных PostgreSQL
с использованием библиотеки gorm
*/
package db

import (
	"SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Database struct представляет абстрактную структуру базы данных
type Database struct {
	SchoolServerDB *gorm.DB
	Logger         *log.Logger
}

// User struct представляет структуру записи пользователя
type User struct {
	gorm.Model
	Login      string `sql:"size:255;unique;index"`
	Password   string `sql:"size:255"`
	Permission bool   `sql:"DEFAULT:false"`
	School     *School
}

// School struct представляет структуру записи школы
type School struct {
	gorm.Model
	Name       string `sql:"size:255;unique"`
	Address    string `sql:"size:255;unique"`
	Permission bool   `sql:"DEFAULT:true"`
}

// NewDatabase создает Database и возвращает указатель на неё
func NewDatabase(logger *log.Logger) (*Database, error) {
	// Подключение к базе данных
	sdb, err := gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=schoolserverdb sslmode=disable")
	if err != nil {
		return nil, err
	}
	// Если таблицы с пользователями не существует, создадим её
	if !sdb.HasTable(&User{}) {
		logger.Info("Table 'users' doesn't exist! Creating new one")
		err := sdb.CreateTable(&User{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("Successfully created 'users' table")
	} else {
		logger.Info("Table 'users' exists")
	}
	// Если таблицы со школами не существует, создадим её
	if !sdb.HasTable(&School{}) {
		logger.Info("'schools' table doesn't exist! Creating new one")
		err := sdb.CreateTable(&School{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("Successfully created 'schools' table")
	} else {
		logger.Info("Table 'schools' exists")
	}
	var count int
	err = sdb.Table("schools").Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count == 0 {
		logger.Info("Table 'schools' empty. Adding our schools")
		// Добавим записи поддерживаемых школ
		school1 := School{
			Address: "http://62.117.74.43/", Name: "Европейская гимназия",
		}
		err = sdb.Save(&school1).Error
		if err != nil {
			return nil, err
		}
		logger.Info("Successfully added our schools")
	}

	return &Database{SchoolServerDB: sdb, Logger: logger}, nil
}

// UpdateUser обновляет данные о пользователе
func (db *Database) UpdateUser(login string, passkey string, id int) error {
	var (
		school School
		user   User
	)
	// Получаем запись школы
	err := db.SchoolServerDB.First(&school, id).Error
	if err != nil {
		return err
	}
	// Получаем и обновляем запись пользователя, если не существует - создаем
	where := User{Login: login}
	attrs := User{Password: passkey, School: &school}
	err = db.SchoolServerDB.Where(where).Attrs(attrs).FirstOrCreate(&user).Error
	return err
}

// GetUserAuthData возвращает данные для повторной удаленной авторизации пользователя
func (db *Database) GetUserAuthData(userName string) (*configParser.School, error) {
	var user User
	// Получаем пользователя по логину
	where := User{Login: userName}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	return &configParser.School{Link: user.School.Address, Login: userName, Password: user.Password}, err
}

// GetUserPermission проверяет разрешение пользователя на работу с сервисом
func (db *Database) GetUserPermission(userName string) (bool, error) {
	var user User
	// Получаем пользователя по логину
	where := User{Login: userName}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return false, err
	}
	return user.Permission, nil
}

// GetSchoolPermission проверяет разрешение школы на работу с сервисом
func (db *Database) GetSchoolPermission(id int) (bool, error) {
	var school School
	// Получаем школу по primary key
	err := db.SchoolServerDB.First(&school, id).Error
	if err != nil {
		return false, err
	}
	return school.Permission, nil
}

// GetSchools возвращает информацию о всех поддерживаемых школах.
func (db *Database) GetSchools() ([]School, error) {
	var schools []School
	err := db.SchoolServerDB.Find(&schools).Error
	return schools, err
}
