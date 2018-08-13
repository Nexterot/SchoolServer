/*
Package db содержит необходимое API для работы с базой данных PostgreSQL с использованием библиотеки gorm
*/
package db

import (
	"fmt"
	"strconv"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions/session"

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

// Database struct представляет абстрактную структуру базы данных
type Database struct {
	SchoolServerDB *gorm.DB
	Logger         *log.Logger
}

// School struct представляет структуру записи школы
type School struct {
	gorm.Model
	Name       string `sql:"size:255;unique"`
	Initials   string `sql:"size:255"`
	Address    string `sql:"size:255;unique"`
	Permission bool   `sql:"DEFAULT:true"`
	Users      []User // has-many relation
}

// User struct представляет структуру записи пользователя
type User struct {
	gorm.Model
	SchoolID   uint      // parent id
	IsParent   bool      `sql:"DEFAULT:false"`
	Login      string    `sql:"size:255;index"`
	Password   string    `sql:"size:255"`
	Permission bool      `sql:"DEFAULT:false"`
	Students   []Student // has-many relation
}

// Student struct представляет структуру ученика
type Student struct {
	gorm.Model
	UserID      uint   // parent id
	NetSchoolID int    // id ученика в системе NetSchool
	ClassID     string // id класса, в котором учится ученик
	Name        string `sql:"size:255"`
	Days        []Day  // has-many relation
}

// Day struct представляет структуру дня с дз
type Day struct {
	gorm.Model
	StudentID uint   // parent id
	Date      string `sql:"size:255"`
	Tasks     []Task // has-many relation
}

// Task представляет структуру задания
type Task struct {
	gorm.Model
	DayID      uint // parent id
	LessonID   int  // id урока (CID)
	HometaskID int  // id домашнего задания (AID)
	TP         int  // TP
	Status     int  // новое\просмотренное\выполненное
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
		// Student
		logger.Info("DB: Creating 'students' table")
		err = sdb.CreateTable(&Student{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'students' table")
		// Day
		logger.Info("DB: Creating 'day' table")
		err = sdb.CreateTable(&Day{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'day' table")
		// Task
		logger.Info("DB: Creating 'task' table")
		err = sdb.CreateTable(&Task{}).Error
		if err != nil {
			return nil, err
		}
		logger.Info("DB: Successfully created 'task' table")
	} else {
		logger.Info("DB: Table 'users' exists")
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
			Address: "62.117.74.43", Name: "Европейская гимназия", Initials: "ЕГ",
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
func (db *Database) UpdateUser(login string, passkey string, isParent bool, schoolID int, childrenMap map[string]ss.Student) error {
	var (
		school  School
		student Student
		user    User
	)
	// Получаем запись школы
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting school with id='%d'", schoolID)
	}
	// Получаем запись пользователя
	where := User{Login: login, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
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
			user := User{
				SchoolID: uint(schoolID),
				IsParent: isParent,
				Login:    login,
				Password: passkey,
				Students: students,
			}
			errInner := db.SchoolServerDB.Create(&user).Error
			if errInner != nil {
				return errors.Wrapf(errInner, "Error creating user='%v'", user)
			}
			return nil
		}
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Пользователь найден, обновим
	user.Password = passkey
	user.SchoolID = uint(schoolID)
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving updated user='%v'", user)
	}
	return nil
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
		return nil, errors.Wrapf(err, "Error query school by id='%d'", schoolID)
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
		return errors.Wrapf(fmt.Errorf("record not found"), "No student with id='%d' found for userName='%s'", studentID, userName)
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
				if task.AID == dbTask.HometaskID {
					dbTaskFound = true
					newTask = dbTask
					break
				}
			}
			if !dbTaskFound {
				// Задания не существует, надо создать
				newTask = Task{DayID: newDay.ID, LessonID: task.CID, HometaskID: task.AID, Status: StatusTaskNew, TP: task.TP}
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

// GetTaskInfo получает информацию о задании - дату, CID, TP, id, classID ученика
func (db *Database) GetTaskInfo(userName string, schoolID int, taskID int) (string, int, int, int, string, error) {
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
		return "", -1, -1, -1, "", errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем таски с таким taskID
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		return "", -1, -1, -1, "", errors.Wrapf(err, "Error query tasks='%v'", w)
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			return "", -1, -1, -1, "", errors.Wrapf(err, "Error query day by id='%d'", t.DayID)
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			return "", -1, -1, -1, "", errors.Wrapf(err, "Error query student by day id='%d'", day.StudentID)
		}

		// Если совпал id пользователя, обновить статус и вернуть дату
		if user.ID == student.UserID {
			if t.Status == StatusTaskNew {
				t.Status = StatusTaskSeen
				err = db.SchoolServerDB.Save(&t).Error
				if err != nil {
					return "", -1, -1, -1, "", errors.Wrapf(err, "Error saving updated task='%v'", t)
				}
			}
			return day.Date, t.LessonID, t.TP, student.NetSchoolID, student.ClassID, nil
		}
	}
	// Таск не найден
	return "", -1, -1, -1, "", fmt.Errorf("record not found")
}

// TaskMarkDone меняет статус задания на "Выполненное"
func (db *Database) TaskMarkDone(userName string, schoolID int, taskID int) error {
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
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		return errors.Wrapf(err, "Error query tasks with id='%v'", taskID)
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query day with id='%d'", t.DayID)
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student).Error
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
func (db *Database) TaskMarkUndone(userName string, schoolID int, taskID int) error {
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
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		return errors.Wrapf(err, "Error query tasks with id='%v'", taskID)
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

// GetStudents возвращает список учеников пользователя
func (db *Database) GetStudents(userName string, schoolID int) (map[string]int, error) {
	var (
		user     User
		students []Student
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем ассоциированных с пользователем учеников
	err = db.SchoolServerDB.Model(&user).Related(&students).Error
	if err != nil {
		return nil, errors.Wrapf(err, "Error query user='%v' students", user)
	}
	// Заполним ответ
	childrenMap := make(map[string]int)
	for _, st := range students {
		childrenMap[st.Name] = st.NetSchoolID
	}
	return childrenMap, nil
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
