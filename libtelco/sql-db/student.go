// student
package db

import (
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

// Student struct представляет структуру ученика
type Student struct {
	gorm.Model
	UserID      uint   // parent id
	Name        string `sql:"size:255"`
	NetSchoolID int    // id ученика в системе NetSchool
	ClassID     string // id класса, в котором учится ученик
	Days        []Day  // has-many relation
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
