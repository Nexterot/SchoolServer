package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"fmt"
)

type User struct {
	gorm.Model
	Login string
	Password string
	Name string
	School string
	UserType int //1 школьник 2 родителб
	Children []User
}

func addUser(user User,db *gorm.DB) bool {
	suc:=db.NewRecord(user)

	if !suc  {
		return !suc
	}
	db.Create(&user)
	return true
	}

func findUserById (id uint, db *gorm.DB) *User {
	var user User
	db.First(&user,id)
	return &user
}

func findUserByLogin (login string,db *gorm.DB) *User {
	var user User
	db.Where("login = ?", "login").First(&user)
	return &user
}

func UpdateUser(user *User,relation string,newVal string,db *gorm.DB) {
	db.Model(user).Update(relation,newVal)
}

func UpdateUserById (id uint,relation string, newVal string,db *gorm.DB){
	UpdateUser(findUserById(id,db),relation,newVal,db)
}
func UpdateUserByLogin (login string,relation string, newVal string,db *gorm.DB){
	UpdateUser(findUserByLogin(login,db),relation,newVal,db)
}


func deleteUser(user *User,db *gorm.DB) {
	db.Delete(user)
}


func main() {

	db, err := gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=users")
	if nil!=err  {
		panic(err)

	}

	defer db.Close()
	db.DropTableIfExists(&User{}, "users")
	db.CreateTable(&User{})

	user:=User{Login:"login", Password:"password", Name:"имя1",School: "gymnasy #228", UserType: 1}

	addUser(user,db)
	user1:=findUserByLogin("login",db)
	fmt.Println(user1.Name)
	UpdateUserById(1,"name","имя2",db)
	user1=findUserByLogin("login",db)
	fmt.Println(user1.Name)
	deleteUser(user1,db)

}
