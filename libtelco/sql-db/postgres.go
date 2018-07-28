/*
Package db содержит необходимое API для работы с базой данных PostgreSQL с использованием библиотеки gorm
*/
package db

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	dt "SchoolServer/libtelco/sessions/data-types"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
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
			Address: "http://62.117.74.43/", Name: "Европейская гимназия",
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
func (db *Database) UpdateUser(login string, passkey string, isParent bool, schoolID int, childrenMap map[string]string) error {
	var (
		school  School
		student Student
		user    User
	)
	// Получаем запись школы
	err := db.SchoolServerDB.First(&school, schoolID).Error
	if err != nil {
		db.Logger.Info("DB: Error getting school by id", "SchoolID", schoolID)
		return err
	}
	// Получаем запись пользователя
	where := User{Login: login, SchoolID: uint(schoolID)}
	err = db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			// Пользователь не найден, создадим
			students := make([]Student, len(childrenMap))
			i := 0
			for childName, childID := range childrenMap {
				id, errInner := strconv.Atoi(childID)
				if errInner != nil {
					db.Logger.Info("DB: Error occured when converting IDs from childrenMap", "ChildID", childID)
					return errInner
				}
				student = Student{Name: childName, NetSchoolID: id, Days: []Day{}}
				errInner = db.SchoolServerDB.Create(&student).Error
				if errInner != nil {
					db.Logger.Info("DB: Error occured when creating student for user", "User", user, "Student", student)
					return errInner
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
				db.Logger.Error("DB: Error occured when creating user", "User", user)
			}
			return errInner
		}
		db.Logger.Error("DB: Error occured when getting user from db", "User params", where)
		return err
	}
	// Пользователь найден, обновим
	user.Password = passkey
	user.SchoolID = uint(schoolID)
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when saving updated data for user", "User", user)
	}
	return err
}

// GetUserAuthData возвращает данные для повторной удаленной авторизации пользователя
func (db *Database) GetUserAuthData(userName string, schoolID int) (*cp.School, error) {
	var user User
	// Получаем пользователя по школе и логину
	where := User{Login: userName}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	return &cp.School{Link: "user.School.Address", Login: userName, Password: user.Password}, err
}

// GetUserPermission проверяет разрешение пользователя на работу с сервисом
func (db *Database) GetUserPermission(userName string, schoolID int) (bool, error) {
	var user User
	// Получаем пользователя по логину
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting permission for user", "User params", where)
		return false, err
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
		db.Logger.Info("DB: Error occured getting user for updating tasks status", "User params", where)
		return err
	}
	// Получаем ученика по studentID
	err = db.SchoolServerDB.Model(&user).Related(&students).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting user's students list for updating tasks status", "User", user)
		return err
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
		db.Logger.Info("DB: No student found for updating tasks status", "StudentID", studentID)
		return fmt.Errorf("record not found")
	}
	// Получаем список дней у ученика
	err = db.SchoolServerDB.Model(&student).Related(&days).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting student's days list for updating tasks status", "Student", student)
		return err
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
				db.Logger.Info("DB: Error occured when saving new day for updating tasks status", "newDay", newDay)
				return err
			}
			days = append(days, newDay)
		}
		// Получаем список заданий для дня
		err = db.SchoolServerDB.Model(&newDay).Related(&tasks).Error
		if err != nil {
			db.Logger.Info("DB: Error getting tasks list for updating tasks status", "newDAy", newDay)
			return err
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
					db.Logger.Info("DB: Error occured when creating new task for updating tasks status", "newTask", newTask)
					return err
				}
				tasks = append(tasks, newTask)
			}
			// Присвоить статусу таска из пакета статус таска из БД
			week.Data[dayNum].Lessons[taskNum].Status = newTask.Status
		}
		err = db.SchoolServerDB.Save(&newDay).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when saving new day", "newDay", newDay)
			return err
		}
	}
	err = db.SchoolServerDB.Save(&student).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when saving student", "Student", student)
	}
	return err
}

// GetTaskInfo получает информацию о задании - дату, CID, TP, id ученика
func (db *Database) GetTaskInfo(userName string, schoolID int, taskID int) (string, int, int, int, error) {
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
		db.Logger.Info("DB: Error occured when getting user for getting task info", "User params", where)
		return "", -1, -1, -1, err
	}
	// Получаем таски с таким taskID
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting tasks for getting task info", "Task params", w)
		return "", -1, -1, -1, err
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting days for getting task info", "DayID", t.DayID)
			return "", -1, -1, -1, err
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting student for getting task info", "StudentID", day.StudentID)
			return "", -1, -1, -1, err
		}

		// Если совпал id пользователя, обновить статус и вернуть дату
		if user.ID == student.UserID {
			if t.Status == StatusTaskNew {
				t.Status = StatusTaskSeen
				err = db.SchoolServerDB.Save(&t).Error
				if err != nil {
					db.Logger.Error("DB: Error occured when saving updated task status", "Task", t)
					return "", -1, -1, -1, err
				}
			}
			return day.Date, t.LessonID, t.TP, student.NetSchoolID, nil
		}
	}
	// Таск не найден
	db.Logger.Info("DB: Error occured when searching for task when getting task info")
	return "", -1, -1, -1, fmt.Errorf("record not found")
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
		db.Logger.Info("DB: Error occured when getting user for updating tasks status", "User params", where)
		return err
	}
	// Получаем таски с таким taskID
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting tasks for updating tasks status", "Task params", w)
		return err
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting days for updating tasks status", "DayID", t.DayID)
			return err
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting student for updating tasks status", "StudentID", day.StudentID)
			return err
		}

		// Если совпал id пользователя - поменять статус, сохранить и закончить цикл
		if user.ID == student.UserID {
			t.Status = StatusTaskDone
			err = db.SchoolServerDB.Save(&t).Error
			if err != nil {
				db.Logger.Error("DB: Error occured when saving updated task status", "Task", t)
				return err
			}
			return nil
		}
	}
	// Таск не найден
	db.Logger.Info("DB: Error occured when searching for task to update status")
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
		db.Logger.Info("DB: Error occured when getting user for updating tasks status", "User params", where)
		return err
	}
	// Получаем таски с таким taskID
	w := Task{HometaskID: taskID}
	err = db.SchoolServerDB.Where(w).Find(&tasks).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting tasks for updating tasks status", "TaskID", taskID)
		return err
	}
	// Найдем нужный таск
	for _, t := range tasks {
		// Получим день по DayID
		err = db.SchoolServerDB.First(&day, t.DayID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting days for updating tasks status", "DayID", t.DayID)
			return err
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			db.Logger.Info("DB: Error occured when getting student for updating tasks status", "StudentID", day.StudentID)
			return err
		}
		// Если совпал id пользователя - поменять статус, сохранить и закончить цикл
		if user.ID == student.UserID {
			t.Status = StatusTaskSeen
			err = db.SchoolServerDB.Save(&t).Error
			if err != nil {
				db.Logger.Error("DB: Error occured when saving updated task status", "Task", t)
			}
			return err
		}
	}
	// Таск не найден
	db.Logger.Info("DB: Error occured when searching for task to update status")
	return fmt.Errorf("record not found")
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
		db.Logger.Info("DB: Error occured when getting user for getting childrenMap", "User params", where)
		return nil, err
	}
	// Получаем ассоциированных с пользователем учеников
	err = db.SchoolServerDB.Model(&user).Related(&students).Error
	if err != nil {
		db.Logger.Info("DB: Error occured when getting students list for getting childrenMap", "User", user)
		return nil, err
	}
	// Заполним ответ
	childrenMap := make(map[string]int)
	for _, st := range students {
		childrenMap[st.Name] = st.NetSchoolID
	}
	return childrenMap, nil
}
