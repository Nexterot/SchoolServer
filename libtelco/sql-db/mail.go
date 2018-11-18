// mail
package db

import (
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

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
			if post.ID == dbPost.NetschoolID && section == dbPost.Section {
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
