// forum
package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/pkg/errors"
)

// ForumTopic struct представляет структуру темы на форуме
type ForumTopic struct {
	gorm.Model
	UserID      uint // parent id
	NetschoolID int
	Date        string
	Creator     string
	Title       string
	Unread      bool
	Answers     int
	Posts       []ForumPost // has-many relation
}

// ForumPost struct представляет структуру сообщения в теме на форуме
type ForumPost struct {
	gorm.Model
	ForumTopicID uint // parent id
	Date         string
	Author       string
	Message      string
	Unread       bool
}

// UpdateTopicsStatuses добавляет в БД несуществующие темы форума и обновляет статусы
func (db *Database) UpdateTopicsStatuses(userName string, schoolID int, themes *dt.ForumThemesList) error {
	var (
		user     User
		newTopic ForumTopic
		topics   []ForumTopic
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем список тем у пользователя
	err = db.SchoolServerDB.Model(&user).Related(&topics).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' forum topics", user)
	}
	// Гоняем по темам из пакета
	for topicNum, topic := range themes.Posts {
		// Найдем подходящую тему в БД
		topicFound := false
		for _, dbTopic := range topics {
			if topic.ID == dbTopic.NetschoolID {
				topicFound = true
				newTopic = dbTopic
				break
			}
		}
		if !topicFound {
			// Темы не существует, надо создать
			newTopic = ForumTopic{UserID: user.ID, NetschoolID: topic.ID, Date: topic.Date, Creator: topic.Creator, Title: topic.Title, Unread: true, Answers: topic.Answers, Posts: []ForumPost{}}
			err = db.SchoolServerDB.Create(&newTopic).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newTopic='%v'", newTopic)
			}
			topics = append(topics, newTopic)
		}
		// Присвоить статусу темы из пакета статус темы из БД
		themes.Posts[topicNum].Unread = newTopic.Unread
	}
	// Сохраним пользователя
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}

// UpdatePostsStatuses добавляет в БД несуществующие сообщения форума и обновляет статусы
func (db *Database) UpdatePostsStatuses(userName string, schoolID int, themeID int, topics *dt.ForumThemeMessages) error {
	var (
		user       User
		topic      ForumTopic
		newMessage ForumPost
		messages   []ForumPost
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем нужную тему у пользователя
	wh := ForumTopic{NetschoolID: themeID, UserID: user.ID}
	err = db.SchoolServerDB.Where(wh).First(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting forum topic='%v'", wh)
	}
	// Помечаем тему как прочитанную
	topic.Unread = false
	// Получаем сообщения у темы
	err = db.SchoolServerDB.Model(&topic).Related(&messages).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting topic='%v' messages", topic)
	}
	// Гоняем по сообщениям из пакета
	for postNum, post := range topics.Messages {
		// Найдем подходящее сообщение в БД
		postFound := false
		for _, dbPost := range messages {
			if post.Date == dbPost.Date && post.Author == dbPost.Author && post.Message == dbPost.Message {
				postFound = true
				newMessage = dbPost
				break
				// В этом случае гарантируется, что сообщение уже было прочитано
			}
		}
		if !postFound {
			// Сообщения не существует, надо создать
			newMessage = ForumPost{ForumTopicID: topic.ID, Date: post.Date, Author: post.Author, Unread: false, Message: post.Message}
			err = db.SchoolServerDB.Create(&newMessage).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newMessage='%v'", newMessage)
			}
			messages = append(messages, newMessage)
			// Присвоить статусу сообщения из пакет из БД "не прочитано"
			topics.Messages[postNum].Unread = true
		}
	}
	// Сохраним тему
	err = db.SchoolServerDB.Save(&topic).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving topic='%v'", topic)
	}
	return nil
}
