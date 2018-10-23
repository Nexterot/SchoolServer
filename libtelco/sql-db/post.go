// post
package db

import (
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

// Post struct представляет структуру объявления
type Post struct {
	gorm.Model
	UserID   uint // parent id
	Unread   bool
	Author   string
	Title    string
	Date     string
	Message  string
	File     string
	FileName string
}

// UpdatePosts добавляет в БД несуществующие объявления и обновляет их статусы
func (db *Database) UpdatePosts(userName string, schoolID int, ps *dt.Posts) error {
	var (
		user    User
		newPost Post
		posts   []Post
	)
	// Получаем пользователя по логину и schoolID
	where := User{Login: userName, SchoolID: uint(schoolID)}
	err := db.SchoolServerDB.Where(where).First(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error query user='%v'", where)
	}
	// Получаем список объявлений у пользователя
	err = db.SchoolServerDB.Model(&user).Related(&posts).Error
	if err != nil {
		return errors.Wrapf(err, "Error getting user='%v' posts", user)
	}
	// Гоняем по объявлениям из пакета
	for _, post := range ps.Posts {
		// Найдем подходящую тему в БД
		postFound := false
		for _, dbPost := range posts {
			if post.Author == dbPost.Author && post.Title == dbPost.Title && post.Date == dbPost.Date {
				postFound = true
				newPost = dbPost
				break
			}
		}
		if !postFound {
			// Объявления не существует, надо создать
			newPost = Post{UserID: user.ID, Unread: post.Unread, Author: post.Author, Title: post.Title, Date: post.Date, Message: post.Message, File: post.FileLink, FileName: post.FileName}
			err = db.SchoolServerDB.Create(&newPost).Error
			if err != nil {
				return errors.Wrapf(err, "Error creating newPost='%v'", newPost)
			}
			posts = append(posts, newPost)
		}
	}
	// Сохраним пользователя
	err = db.SchoolServerDB.Save(&user).Error
	if err != nil {
		return errors.Wrapf(err, "Error saving user='%v'", user)
	}
	return nil
}
