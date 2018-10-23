// user
package db

import (
	"strconv"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
	"github.com/pkg/errors"
)

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
	MailMessages []MailMessage // has-many relation
	Posts        []Post        // has-many realtion
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
