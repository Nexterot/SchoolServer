package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"SchoolServer/libtelco/config-parser"
	"strconv"
	"errors"
	"fmt"
	"log"
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

type school struct {
	Name 	string `json:"name"`
	Id 		string `json:"id"`
	Website string `json:"website"`
}

type SchoolList struct {
	Schools []school `json:"schools"`
}


func NewDatabase(log* log.Logger) (*Database,error) {//создает новую структуру Database и возвращает указатель на неё.
	//открываем дб
	sdb,err :=gorm.Open("postgres", "host=localhost port=5432 user=test_user password=qwerty dbname=schoolserverdb sslmode=disable")
	log.Println("Connecting to DB")
	if err!=nil {
		return &Database{},errors.New("Can't connect to database")
	}
	if !sdb.HasTable(&User{}){ //добавляем таблицу с пользователями, если её нет
		err:=sdb.CreateTable(&User{}).Error
		log.Println("Creating 'users' table")
		if err!=nil {
			return &Database{},err
		}
	}

	if !sdb.HasTable(&School{}){ //добавляем таблицу со школами, если её нет
		err:=sdb.CreateTable(&School{}).Error
		log.Println("Creating 'schools' table")
		if err!=nil {
			return &Database{},err
		}
	}
	return &Database{SchoolServerDB:sdb},nil
}

func (db *Database) UpdateUser(login string, passkey string, id int,log* log.Logger) (bool,error) {
	//обновляет данные о пользователе. Если пользователя с таким именем нет в БД, создаёт.
	var user User
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",login).First(&user).Error //ищем пользователя в дб по логину
	if err!=nil {
		return false,err
	}
	log.Printf("Finding user %s in DB\n",login)
	if user.ID>0 { //существует
		user.Password=passkey
		user.SchoolID=id
		err=db.SchoolServerDB.Save(&user).Error //обновляем данные
		log.Printf("Updating user %s in DB\n",login)
		if err!=nil {
			return false,err
		}
		return true,nil
	}else { //не существует
		err=db.SchoolServerDB.Create(&User{Login:login,Password:passkey,SchoolID:id,Permission:true}).Error
		log.Printf("Creating user %s in DB\n",login)
		if err!=nil {
			return false,err
		}
		return false,nil
	}
	//возвращаемое значение:
	//true, если пользователь ранее существовал, false иначе
}

func (db *Database) GetUserAuthData(userName string,log* log.Logger) (*configParser.School, error){
	//возвращает данные для повторной авторизации пользователя с именем userName
	var school School
	var user User
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user).Error //ищем пользователя по логину
	log.Printf("Finding user %s in DB\n",userName)
	if err!=nil {
		return &configParser.School{},err
	}
	if user.ID==0 { //нет такого пользователя
		return &configParser.School{},fmt.Errorf("User with name "+userName+" doesn't exist")
	}
	err=db.SchoolServerDB.Find(&School{}).First(&school, user.SchoolID).Error //ищем школу пользователя по ид
	log.Printf("Finding school %d in DB\n",user.SchoolID);
	if err!=nil {
		return &configParser.School{},err
	}
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

func (db *Database) GetPermission(userName string, schoolId int,log* log.Logger) (bool,error) {
	//возвращает permission для пользователя школы с номером schoolId.
	var (
		user User
		school School
	)
	err:=db.SchoolServerDB.Find(&User{}).Where("login = ?",userName).First(&user).Error //ищем пользователя по логину
	log.Printf("Finding user %s in DB",userName)
	if err!=nil {
		return false,err
	}
	err=db.SchoolServerDB.Find(&School{}).First(&school,schoolId).Error //ищем школу по ид
	log.Printf("Finding school %d in DB",schoolId)
	if err!=nil {
		return false,err
	}
	return user.Permission||school.Permission,nil;
}
