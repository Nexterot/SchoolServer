// school
package db

import (
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

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
