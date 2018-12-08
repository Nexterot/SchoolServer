// diary
package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/pkg/errors"
)

// Day struct представляет структуру дня с дз
type Day struct {
	gorm.Model
	StudentID uint     // parent id
	Date      string   `sql:"size:255"`
	Tasks     []Task   // has-many relation
	Lessons   []Lesson // has-many relation
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

// Статусы заданий
const (
	_ = iota
	StatusTaskNew
	StatusTaskSeen
	StatusTaskDone
)

// Lesson struct представляет структуру урока в расписании
type Lesson struct {
	gorm.Model
	DayID     uint // parent id
	Begin     string
	End       string
	Name      string
	Classroom string
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

// TaskMarkSeen меняет статус задания на "Просмотренное"
func (db *Database) TaskMarkSeen(userName string, schoolID int, AID, CID, TP int) error {
	var (
		tasks   []Task
		day     Day
		days    []Day
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
		db.Logger.Info("Getting all days...")
		err = db.SchoolServerDB.Find(&days).Error
		if err != nil {
			return errors.Wrapf(err, "Error query all days")
		}
		for i, d := range days {
			db.Logger.Info("Days", "num", i, "day", d)
		}
		err = db.SchoolServerDB.Where("id = ?", t.DayID).First(&day).Error
		if err != nil {
			return errors.Wrapf(err, "Error query day with id='%v', task='%v'", t.DayID, t)
		}
		// Получим студента по дню
		err = db.SchoolServerDB.First(&student, day.StudentID).Error
		if err != nil {
			return errors.Wrapf(err, "Error query student with id='%v'", day.StudentID)
		}
		// Если совпал id пользователя - поменять статус, сохранить и закончить цикл
		if user.ID == student.UserID {
			if t.Status == StatusTaskNew {
				t.Status = StatusTaskSeen
				err = db.SchoolServerDB.Save(&t).Error
				if err != nil {
					return errors.Wrapf(err, "Error saving updated task='%v'", t)
				}
			}
			return nil
		}
	}
	// Таск не найден
	return fmt.Errorf("task record not found")
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

// UpdateLessons добавляет в БД несуществующие уроки в расписании
func (db *Database) UpdateLessons(userName string, schoolID int, studentID int, week *dt.TimeTable) error {
	var (
		student   Student
		students  []Student
		user      User
		newDay    Day
		days      []Day
		newLesson Lesson
		lessons   []Lesson
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
	for _, day := range week.Days {
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
			newDay = Day{StudentID: student.ID, Date: date, Tasks: []Task{}, Lessons: []Lesson{}}
			err = db.SchoolServerDB.Create(&newDay).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newDay='%v'", newDay)
			}
			days = append(days, newDay)
		}
		// Получаем список уроков для дня
		err = db.SchoolServerDB.Model(&newDay).Related(&lessons).Error
		if err != nil {
			return errors.Wrapf(err, "Error getting newDay='%v' lessons", newDay)
		}
		// Гоняем по урокам
		for _, lesson := range day.Lessons {
			// Найдем подходящий урок в БД
			dbLessonFound := false
			for _, dbLesson := range lessons {
				if lesson.Begin == dbLesson.Begin {
					if lesson.Name != dbLesson.Name || lesson.ClassRoom != dbLesson.Classroom {
						// Произошли изменения
						// по полю Begin сравнивали, они равны, End наверное тоже
						dbLesson.Name = lesson.Name
						dbLesson.Classroom = lesson.ClassRoom
						err = db.SchoolServerDB.Save(&dbLesson).Error
						if err != nil {
							return errors.Wrapf(err, "Error saving updated lesson='%v'", dbLesson)
						}
					}
					// Если урок нашелся, обновим поля в БД
					dbLessonFound = true
					newLesson = dbLesson
					break
				}
			}
			if !dbLessonFound {
				// Урока не существует, надо создать
				if len(lessons) == 1 && lessons[0].Begin == "00:00" {
					// Значит, что выходной перестал быть выходным
					// Псевдоудаляем
					err = db.SchoolServerDB.Delete(&lessons[0]).Error
					if err != nil {
						return errors.Wrapf(err, "Error deleting lesson='%v'", lessons[0])
					}
				}
				newLesson = Lesson{DayID: newDay.ID, Begin: lesson.Begin, End: lesson.End, Name: lesson.Name, Classroom: lesson.ClassRoom}
				err = db.SchoolServerDB.Create(&newLesson).Error
				if err != nil {
					return errors.Wrapf(err, "Error creating newLesson='%v'", newLesson)
				}
				lessons = append(lessons, newLesson)
			}
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
			newDay = Day{StudentID: student.ID, Date: date, Tasks: []Task{}, Lessons: []Lesson{}}
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
				if task.AID == dbTask.AID && task.CID == dbTask.CID && task.TP == dbTask.TP {
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
