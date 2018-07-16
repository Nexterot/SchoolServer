package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"SchoolServer/libtelco/config-parser"
	"strconv"
	"errors"
	"fmt"
)

type Database struct  {
	SchoolServerDB *gorm.DB //указатель на гормовскую дб
}


type User struct {
	gorm.Model
	Login string //логин пользователя
	Password string //зашифрованный пароль пользователя
	SchoolID int //ид школы как в дб для школ
	Permission bool //разрешение
}

type School struct {
	gorm.Model
	Address string //интернет адрес сервера
	Permission bool //разрешение
}

func NewDatabase() *Database {//создает новую структуру Database и возвращает указатель на неё.
	//открываем дб
	sdb,err :=gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=schoolserverdb sslmode=disable")
	if err!=nil {
		panic(err)
	}
	if !sdb.HasTable(&User{}){ //добавляем таблицу с пользователями, если её нет
		sdb.CreateTable(&User{})
	}

	if !sdb.HasTable(&School{}){ //добавляем таблицу со школами, если её нет
		sdb.CreateTable(&School{})
	}
	return &Database{SchoolServerDB:sdb}
}

func (db *Database) UpdateUser(login string, passkey string, id int/*было стринг но я решил: в  интах проще*/) bool {
	//обновляет данные о пользователе. Если пользователя с таким именем нет в БД, создаёт.
	var user User
	db.SchoolServerDB.Find(&User{}).Where("login = ?",login).First(&user) //ищем пользователя в лб по логину
	if user.ID>0 { //существует
		user.Password=passkey
		user.SchoolID=id
		db.SchoolServerDB.Save(&user) //обновляем данные
		return true
	}else { //не существует
		db.SchoolServerDB.Create(&User{Login:login,Password:passkey,SchoolID:id,Permission:true})
		return false
	}
	//возвращаемое значение:
	//true, если пользователь ранее существовал, false иначе
}

func (db *Database) GetUserAuthData(userName string) (*configParser.School, error){
	//возвращает данные для повторной авторизации пользователя с именем userName
	var school School
	var user User
	db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user) //ищем пользователя по логину
	if user.ID==0 { //нет такого пользователя
		return &configParser.School{},fmt.Errorf("User with name "+userName+" doesn't exist")
	}
	db.SchoolServerDB.Find(&School{}).First(&school, user.SchoolID) //ищем школу пользователя по ид
	if school.ID==0 {
		return &configParser.School{},errors.New("School with ID "+strconv.Itoa(user.SchoolID)+" doesn't exist")
	}
	return &configParser.School{Link:school.Address,Login:userName,Password:user.Password,},nil
	//*cp.School –  указатель на структуру School из “SchoolServer/libtelco/config-parser“,
	//у которой обязательно заполнены поля Link, Login, Password;
	// где Link – путь к серверу школы пользователя без указания протокола (!) (например, “62.117.74.43”),
	// Login – имя пользователя, Password – md5-хеш пароля пользователя
	//err – ошибка, если она была, nil иначе
}

func (db *Database) GetPermission(userName string, schoolId int/*было стринг но я решил: в  интах проще*/) bool {
	//возвращает permission для пользователя школы с номером schoolId.
	var (
		user User
		school School
	)
	db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user) //ищем пользователя по логину
	db.SchoolServerDB.Find(&School{}).First(&school,schoolId) //ищем школу по ид
	return user.Permission||school.Permission;
}
