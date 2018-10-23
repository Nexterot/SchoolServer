// device
package db

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
	"github.com/pkg/errors"
)

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

// UpdatePushSettings обновляет настройки предпочтений устройства по поводу push уведомлений
func (db *Database) UpdatePushSettings(userName string, schoolID int, systemType int, token string, marks, tasks int, reports, schedule, mail, forum, resources bool) error {
	var (
		user User
		dev  Device
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем ассоциированное устройство по systemType и token
	wh := Device{Token: token, SystemType: systemType}
	err = db.SchoolServerDB.Where(wh).First(&dev).Error
	if err != nil {
		return errors.Wrapf(err, "Error query device='%v'", wh)
	}
	// Обновляем настройки
	dev.MarksNotification = marks
	dev.TasksNotification = tasks
	dev.ReportsNotification = reports
	dev.ScheduleNotification = schedule
	dev.MailNotification = mail
	dev.ForumNotification = forum
	dev.ResourcesNotification = resources
	// Сохраняем в бд
	err = db.SchoolServerDB.Save(&dev).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving updated device='%v'", dev)
	}
	return nil
}
